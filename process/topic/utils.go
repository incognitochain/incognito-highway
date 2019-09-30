package topic

import "sort"

func isBroadcastMessage(message string) bool {
	if message == CmdBFT {
		return true
	}
	return false
}

func isValidMessage(message string) bool {
	sort.Slice(Message4Process, func(i, j int) bool {
		return Message4Process[i] < Message4Process[j]
	})
	idx := sort.SearchStrings(Message4Process, message)
	if (idx < 0) || (idx >= len(Message4Process)) {
		return false
	}
	return true
}
