package tradeoffer

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type EscrowDuration struct {
	DaysMyEscrow    uint32
	DaysTheirEscrow uint32
}

func parseEscrowDuration(data []byte) (*EscrowDuration, error) {
	// TODO: why we are using case insensitive matching?
	myRegex := regexp.MustCompile("(?i)g_daysMyEscrow[\\s=]+(\\d+);")
	theirRegex := regexp.MustCompile("(?i)g_daysTheirEscrow[\\s=]+(\\d+);")

	myM := myRegex.FindSubmatch(data)
	theirM := theirRegex.FindSubmatch(data)

	if myM == nil || theirM == nil {
		// check if access token is valid
		notFriendsRegex := regexp.MustCompile(">You are not friends with this user<")
		notFriendsM := notFriendsRegex.FindSubmatch(data)
		if notFriendsM == nil {
			return nil, errors.New("regexp does not match")
		} else {
			return nil, errors.New("you are not friends with this user")
		}
	}

	myEscrow, err := strconv.ParseUint(string(myM[1]), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse my duration into uint: %v", err)
	}
	theirEscrow, err := strconv.ParseUint(string(theirM[1]), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse their duration into uint: %v", err)
	}

	return &EscrowDuration{
		DaysMyEscrow:    uint32(myEscrow),
		DaysTheirEscrow: uint32(theirEscrow),
	}, nil
}
