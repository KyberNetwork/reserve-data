package caller

import (
	"runtime"
	"strings"
)

func getFrame(skipFrames int) runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}
	return frame
}

// stripName removes the 2 parts of the fullname
func stripName(functionName string) string {
	splits := strings.Split(functionName, "/")
	if len(splits) >= 3 {
		return strings.Join(splits[2:], "/")
	}
	return functionName
}

// GetCurrentFunctionName returns the full name of the function call it
func GetCurrentFunctionName() string {
	// Skip GetCurrentFunctionName
	return stripName(getFrame(1).Function)
}

// GetCurrentFunctionName returns the full name of the parent function call it
func GetCallerFunctionName() string {
	// Skip GetCallerFunctionName and the function to get the caller of
	return stripName(getFrame(2).Function)
}
