// Code generated by "stringer -type=Deployment -linecomment"; DO NOT EDIT.

package deployment

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Development-0]
	_ = x[Production-1]
	_ = x[Simulation-2]
}

const _Deployment_name = "developproductionsimulation"

var _Deployment_index = [...]uint8{0, 7, 17, 27}

func (i Deployment) String() string {
	if i < 0 || i >= Deployment(len(_Deployment_index)-1) {
		return "Deployment(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Deployment_name[_Deployment_index[i]:_Deployment_index[i+1]]
}
