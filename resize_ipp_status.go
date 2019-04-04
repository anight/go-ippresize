package ippresize

//go:generate stringer -type=IppStatus

/*
#include <ipp.h>

enum {
	ipp_status_ippStsNullPtrErr               = ippStsNullPtrErr,
	ipp_status_ippStsNoOperation              = ippStsNoOperation,
	ipp_status_ippStsSizeErr                  = ippStsSizeErr,
	ipp_status_ippStsExceededSizeErr          = ippStsExceededSizeErr,
	ipp_status_ippStsInterpolationErr         = ippStsInterpolationErr,
	ipp_status_ippStsNoAntialiasing           = ippStsNoAntialiasing,
	ipp_status_ippStsNotSupportedModeErr      = ippStsNotSupportedModeErr,
	ipp_status_ippStsContextMatchErr          = ippStsContextMatchErr,
	ipp_status_ippStsNumChannelsErr           = ippStsNumChannelsErr,
	ipp_status_ippStsBorderErr                = ippStsBorderErr,
	ipp_status_ippStsStepErr                  = ippStsStepErr,
	ipp_status_ippStsOutOfRangeErr            = ippStsOutOfRangeErr,
	ipp_status_ippStsSizeWrn                  = ippStsSizeWrn,
};

*/
import "C"

type IppStatus C.IppStatus

const (
	IppStsNullPtrErr          IppStatus = C.ipp_status_ippStsNullPtrErr
	IppStsNoOperation         IppStatus = C.ipp_status_ippStsNoOperation
	IppStsSizeErr             IppStatus = C.ipp_status_ippStsSizeErr
	IppStsExceededSizeErr     IppStatus = C.ipp_status_ippStsExceededSizeErr
	IppStsInterpolationErr    IppStatus = C.ipp_status_ippStsInterpolationErr
	IppStsNoAntialiasing      IppStatus = C.ipp_status_ippStsNoAntialiasing
	IppStsNotSupportedModeErr IppStatus = C.ipp_status_ippStsNotSupportedModeErr
	IppStsContextMatchErr     IppStatus = C.ipp_status_ippStsContextMatchErr
	IppStsNumChannelsErr      IppStatus = C.ipp_status_ippStsNumChannelsErr
	IppStsBorderErr           IppStatus = C.ipp_status_ippStsBorderErr
	IppStsStepErr             IppStatus = C.ipp_status_ippStsStepErr
	IppStsOutOfRangeErr       IppStatus = C.ipp_status_ippStsOutOfRangeErr
	IppStsSizeWrn             IppStatus = C.ipp_status_ippStsSizeWrn
)
