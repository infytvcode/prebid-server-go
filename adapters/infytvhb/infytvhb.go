package infytvhb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/adapters"
	"github.com/prebid/prebid-server/v3/config"
	"github.com/prebid/prebid-server/v3/errortypes"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

type adapter struct {
	endpoint string
}

// Builder builds a new instance of the Foo adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter, server config.Server) (adapters.Bidder, error) {
	bidder := &adapter{
		endpoint: config.Endpoint,
	}
	return bidder, nil
}

func (a *adapter) MakeRequests(request *openrtb2.BidRequest, requestInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	var requests []*adapters.RequestData
	var errors []error

	headers := http.Header{}
	headers.Add("Content-Type", "application/json;charset=utf-8")
	headers.Add("Accept", "application/json")
	headers.Add("x-openrtb-version", "2.5")

	if request.Device != nil {
		if len(request.Device.UA) > 0 {
			headers.Add("User-Agent", request.Device.UA)
		}

		if len(request.Device.IPv6) > 0 {
			headers.Add("X-Forwarded-For", request.Device.IPv6)
		}

		if len(request.Device.IP) > 0 {
			headers.Add("X-Forwarded-For", request.Device.IP)
		}
	}

	for _, imp := range request.Imp {
		var endpoint string

		if infyExt, err := getImpressionExt(&imp); err == nil {
			endpoint = infyExt.Base

			reqCopy := *request
			reqCopy.Imp = []openrtb2.Imp{}
			reqCopy.Test = 0
			imp.Ext = nil
			// imp.PMP = nil
			imp.BidFloor = infyExt.Floor
			reqCopy.Imp = append(reqCopy.Imp, imp)
			reqCopy.Ext = nil
			requestJSON, err := json.Marshal(reqCopy)
			requestJSON = *(dynamicConfigMerge(infyExt.DspConfigs, &requestJSON, &endpoint))
			fmt.Printf("requestJSON: %v\n", string(requestJSON))
			// dynamic config map
			if err != nil {
				errors = append(errors, err)
				continue
			}
			if infyExt.EndpointType == "VAST_URL" || infyExt.EndpointType == "GAM" {
				requestData := &adapters.RequestData{
					Method: "GET",
					Uri:    endpoint,
				}
				requests = append(requests, requestData)
			} else {
				requestData := &adapters.RequestData{
					Method:  "POST",
					Uri:     endpoint,
					Body:    requestJSON,
					Headers: headers,
				}
				requests = append(requests, requestData)
			}
		}

	}
	return requests, errors
}

func (a *adapter) MakeBids(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	if response.StatusCode == http.StatusNoContent {
		return nil, nil
	} else if response.StatusCode == http.StatusBadRequest {
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode),
		}}
	} else if response.StatusCode != http.StatusOK {
		return nil, []error{&errortypes.BadServerResponse{
			Message: fmt.Sprintf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode),
		}}
	}

	if len(internalRequest.Imp) > 0 {
		var bidResp openrtb2.BidResponse
		impression := &internalRequest.Imp[0]
		if infyExt, err := getImpressionExt(impression); err == nil {
			if infyExt.EndpointType == "VAST_URL" {
				bidResp = openrtb2.BidResponse{
					ID: internalRequest.ID,
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								//TODO: update this by parsing VAST
								{
									ID:    internalRequest.ID,
									AdM:   string(response.Body),
									Price: infyExt.Floor,
									ImpID: internalRequest.Imp[0].ID,
									CID:   "-",
									CrID:  "-",
								},
							},
						},
					},
				}
			} else {
				if err := json.Unmarshal(response.Body, &bidResp); err != nil {
					return nil, []error{err}
				}
				for i, sb := range bidResp.SeatBid {
					for j := range sb.Bid {
						b := &bidResp.SeatBid[i].Bid[j]
						if b.CID == "" {
							b.CID = "-"
						}
						if b.CrID == "" {
							b.CrID = "-"
						}
					}
				}
			}
		}
		bidsCapacity := 1
		if len(bidResp.SeatBid) > 0 {
			bidsCapacity = len(bidResp.SeatBid[0].Bid)
		}
		bidResponse := adapters.NewBidderResponseWithBidsCapacity(bidsCapacity)

		for _, sb := range bidResp.SeatBid {
			for i := range sb.Bid {
				if bidType, err := getMediaTypeForBid(&sb.Bid[i]); err == nil {
					// resolveMacros(&sb.Bid[i])
					bidResponse.Bids = append(bidResponse.Bids, &adapters.TypedBid{
						Bid:     &sb.Bid[i],
						BidType: bidType,
					})
				}
			}
		}

		return bidResponse, nil
	}
	return nil, nil
}

// getMediaTypeForBid determines which type of bid.
func getMediaTypeForBid(bid *openrtb2.Bid) (openrtb_ext.BidType, error) {
	return openrtb_ext.BidTypeVideo, nil
}

// resolveMacros resolves OpenRTB macros in nurl and adm
// func resolveMacros(bid *openrtb2.Bid) {
// 	if bid == nil {
// 		return
// 	}
// 	price := strconv.FormatFloat(bid.Price, 'f', -1, 64)
// 	bid.NURL = strings.Replace(bid.NURL, "${AUCTION_PRICE}", price, -1)
// 	bid.AdM = strings.Replace(bid.AdM, "${AUCTION_PRICE}", price, -1)
// }

// getImpressionExt parses and return first imp ext or nil
func getImpressionExt(imp *openrtb2.Imp) (*openrtb_ext.ExtInfytvHb, error) {
	var bidderExt adapters.ExtImpBidder
	if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
		return nil, &errortypes.BadInput{
			Message: err.Error(),
		}
	}

	var extImpInfyTV openrtb_ext.ExtInfytvHb
	if err := json.Unmarshal(bidderExt.Bidder, &extImpInfyTV); err != nil {
		return nil, &errortypes.BadInput{
			Message: err.Error(),
		}
	}
	return &extImpInfyTV, nil
}

func dynamicConfigMerge(dcs []openrtb_ext.DspConfig, request *[]byte, endpoint *string) *[]byte {
	for _, dc := range dcs {
		switch dc.ConditionType {
		case "geo":
			continue
		case "bundles":
			request = dynamicConfigMergeBundles(dc.Conditions, request, endpoint)
		default:
			continue
		}
	}
	return request
}

func dynamicConfigMergeBundles(Conditions map[string][]openrtb_ext.DspConfigValueEntry, request *[]byte, endpoint *string) *[]byte {
	siteDomain, _ := jsonparser.GetString(*request, "site", "domain")
	appBundleID, _ := jsonparser.GetString(*request, "app", "bundle")

	for bundleId, con := range Conditions {
		if (siteDomain != "" && bundleId == "site") || (siteDomain == bundleId) {
			request = dynamicConfigMergeValues(con, request, endpoint, "site")
		} else if appBundleID == bundleId {
			request = dynamicConfigMergeValues(con, request, endpoint, "app")
		}
	}
	return request
}

func ConvertPath(jsonPath string) []string {
	// Regular expression to capture keys and array indices separately
	re := regexp.MustCompile(`\w+|\[\d+\]`)

	// Find all matches in the given path
	matches := re.FindAllString(jsonPath, -1)

	return matches
}

func dynamicConfigMergeValues(con []openrtb_ext.DspConfigValueEntry, requestJson *[]byte, endpoint *string, t string) *[]byte {
	for _, v := range con {
		switch v.Place {
		case "body":
			for key, value := range v.KeyMap {
				dynamicVal := strings.HasPrefix(value, "Vx")
				if dynamicVal {
					value = strings.TrimPrefix(value, "Vx")
					dynamicKeys := ConvertPath(value)
					var err error
					value, err = jsonparser.GetUnsafeString(*requestJson, dynamicKeys...)
					if err != nil {
						continue
					}
				}
				keys := ConvertPath(key)
				if (t == "app" && keys[0] == "site") || (t == "site" && keys[0] == "app") {
					continue
				}
				if newJson, err := jsonparser.Set(*requestJson, []byte(value), keys...); err == nil {
					requestJson = &newJson
				}
			}
		case "queryparam":
			continue
		case "url":
			continue
		default:
			continue
		}
	}
	return requestJson
}
