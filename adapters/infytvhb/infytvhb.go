package infytvhb

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
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

	for _, imp := range request.Imp {
		var endpoint string

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
	}

	if response.StatusCode == http.StatusBadRequest {
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode),
		}}
	}

	if response.StatusCode != http.StatusOK {
		return nil, []error{&errortypes.BadServerResponse{
			Message: fmt.Sprintf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode),
		}}
	}
	var bidResp openrtb2.BidResponse
	if infyExt, err := getImpressionExt(&internalRequest.Imp[0]); err == nil {
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
				for j, b := range sb.Bid {
					if b.CID == "" {
						bidResp.SeatBid[i].Bid[j].CID = "-"
					}
					if b.CrID == "" {
						bidResp.SeatBid[i].Bid[j].CrID = "-"
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
