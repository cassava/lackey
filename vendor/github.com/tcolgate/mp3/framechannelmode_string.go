// generated by stringer -type=FrameChannelMode; DO NOT EDIT

package mp3

import "fmt"

const _FrameChannelMode_name = "StereoJointStereoDualChannelSingleChannel"

var _FrameChannelMode_index = [...]uint8{6, 17, 28, 41}

func (i FrameChannelMode) String() string {
	if i >= FrameChannelMode(len(_FrameChannelMode_index)) {
		return fmt.Sprintf("FrameChannelMode(%d)", i)
	}
	hi := _FrameChannelMode_index[i]
	lo := uint8(0)
	if i > 0 {
		lo = _FrameChannelMode_index[i-1]
	}
	return _FrameChannelMode_name[lo:hi]
}
