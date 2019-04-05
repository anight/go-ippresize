// Code generated by "stringer -type=IppStatus"; DO NOT EDIT.

package ippresize

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[IppStsNullPtrErr - -8]
	_ = x[IppStsNoOperation-1]
	_ = x[IppStsSizeErr - -6]
	_ = x[IppStsExceededSizeErr - -232]
	_ = x[IppStsInterpolationErr - -23]
	_ = x[IppStsNoAntialiasing-46]
	_ = x[IppStsNotSupportedModeErr - -14]
	_ = x[IppStsContextMatchErr - -13]
	_ = x[IppStsNumChannelsErr - -53]
	_ = x[IppStsBorderErr - -225]
	_ = x[IppStsStepErr - -16]
	_ = x[IppStsOutOfRangeErr - -11]
	_ = x[IppStsSizeWrn-48]
}

const _IppStatus_name = "IppStsExceededSizeErrIppStsBorderErrIppStsNumChannelsErrIppStsInterpolationErrIppStsStepErrIppStsNotSupportedModeErrIppStsContextMatchErrIppStsOutOfRangeErrIppStsNullPtrErrIppStsSizeErrIppStsNoOperationIppStsNoAntialiasingIppStsSizeWrn"

var _IppStatus_map = map[IppStatus]string{
	-232: _IppStatus_name[0:21],
	-225: _IppStatus_name[21:36],
	-53:  _IppStatus_name[36:56],
	-23:  _IppStatus_name[56:78],
	-16:  _IppStatus_name[78:91],
	-14:  _IppStatus_name[91:116],
	-13:  _IppStatus_name[116:137],
	-11:  _IppStatus_name[137:156],
	-8:   _IppStatus_name[156:172],
	-6:   _IppStatus_name[172:185],
	1:    _IppStatus_name[185:202],
	46:   _IppStatus_name[202:222],
	48:   _IppStatus_name[222:235],
}

func (i IppStatus) String() string {
	if str, ok := _IppStatus_map[i]; ok {
		return str
	}
	return "IppStatus(" + strconv.FormatInt(int64(i), 10) + ")"
}