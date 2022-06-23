package infytvhb

import (
	"encoding/json"
	"fmt"
	"net/http"
<<<<<<< HEAD
=======
	"strconv"
	"strings"
>>>>>>> 309d5d81 (Adding infytvhb)

	"github.com/mxmCherry/openrtb/v15/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type adapter struct {
	endpoint string
}

// Builder builds a new instance of the Adot adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter) (adapters.Bidder, error) {
	bidder := &adapter{
		endpoint: config.Endpoint,
	}
	return bidder, nil
}

// MakeRequests makes the HTTP requests which should be made to fetch bids. infytv
func (a *adapter) MakeRequests(request *openrtb2.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
<<<<<<< HEAD
	var requests []*adapters.RequestData
	var errors []error

	requestCopy := *request
	for _, imp := range request.Imp {
		var endpoint string
=======

	var requests []*adapters.RequestData
	var errors []error
	var endpoint string

	requestCopy := *request
	for _, imp := range request.Imp {
>>>>>>> 309d5d81 (Adding infytvhb)
		requestCopy.Imp = []openrtb2.Imp{imp}

		requestJSON, err := json.Marshal(request)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		headers := http.Header{}
		headers.Add("Content-Type", "application/json;charset=utf-8")
		headers.Add("Accept", "application/json")
		headers.Add("x-openrtb-version", "2.5")

<<<<<<< HEAD
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

		if infyExt, err := getImpressionExt(&imp); err != nil {
			endpoint = "http://dsp.infy.tv/rtb/bids/nexage"
		} else {
			endpoint = infyExt.Base
		}

=======
		if infyExt, err := getImpressionExt(&imp); err != nil {
			endpoint = fmt.Sprintf("%s%s", infyExt.Base, infyExt.Path)
		} else {
			endpoint = ""
		}
		fmt.Printf("endpoint: %v\n", endpoint)
>>>>>>> 309d5d81 (Adding infytvhb)
		requestData := &adapters.RequestData{
			Method: "POST",
			Uri:    endpoint,
			Body:   requestJSON,
		}
		requests = append(requests, requestData)
	}
	return requests, errors
}

// MakeBids unpacks the server's response into Bids.
// The bidder return a status code 204 when it cannot delivery an ad.
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
	if err := json.Unmarshal(response.Body, &bidResp); err != nil {
		return nil, []error{err}
	}

	bidsCapacity := 1
	if len(bidResp.SeatBid) > 0 {
		bidsCapacity = len(bidResp.SeatBid[0].Bid)
	}
	bidResponse := adapters.NewBidderResponseWithBidsCapacity(bidsCapacity)

	for _, sb := range bidResp.SeatBid {
		for i := range sb.Bid {
			if bidType, err := getMediaTypeForBid(&sb.Bid[i]); err == nil {
<<<<<<< HEAD
				// resolveMacros(&sb.Bid[i])
=======
				resolveMacros(&sb.Bid[i])
>>>>>>> 309d5d81 (Adding infytvhb)
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
<<<<<<< HEAD
// func resolveMacros(bid *openrtb2.Bid) {
// 	if bid == nil {
// 		return
// 	}
// 	price := strconv.FormatFloat(bid.Price, 'f', -1, 64)
// 	bid.NURL = strings.Replace(bid.NURL, "${AUCTION_PRICE}", price, -1)
// 	bid.AdM = strings.Replace(bid.AdM, "${AUCTION_PRICE}", price, -1)
// }

// getImpressionExt parses and return first imp ext or nil
func getImpressionExt(imp *openrtb2.Imp) (*openrtb_ext.ExtInfytvhb, error) {
=======
func resolveMacros(bid *openrtb2.Bid) {
	if bid == nil {
		return
	}
	price := strconv.FormatFloat(bid.Price, 'f', -1, 64)
	bid.NURL = strings.Replace(bid.NURL, "${AUCTION_PRICE}", price, -1)
	bid.AdM = strings.Replace(bid.AdM, "${AUCTION_PRICE}", price, -1)
}

// getImpressionExt parses and return first imp ext or nil
func getImpressionExt(imp *openrtb2.Imp) (*openrtb_ext.ImpExtInfyTvHb, error) {
>>>>>>> 309d5d81 (Adding infytvhb)
	var bidderExt adapters.ExtImpBidder
	if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
		return nil, &errortypes.BadInput{
			Message: err.Error(),
		}
	}
<<<<<<< HEAD

	var extImpInfyTV openrtb_ext.ExtInfytvhb
=======
	var extImpInfyTV openrtb_ext.ImpExtInfyTvHb
>>>>>>> 309d5d81 (Adding infytvhb)
	if err := json.Unmarshal(bidderExt.Bidder, &extImpInfyTV); err != nil {
		return nil, &errortypes.BadInput{
			Message: err.Error(),
		}
	}
<<<<<<< HEAD
=======

>>>>>>> 309d5d81 (Adding infytvhb)
	return &extImpInfyTV, nil
}
