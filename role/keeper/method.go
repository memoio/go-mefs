package keeper

import (
	"strings"

	"github.com/mgutz/ansi"
)

func GetUsers() ([]string, error) {
	if !IsKeeperServiceRunning() {
		return nil, ErrKeeperServiceNotReady
	}
	var res []string
	PInfo.Range(func(uid, groupsInfo interface{}) bool {
		thisuid, ok := uid.(string)
		if !ok {
			return false
		}
		thisGroupsInfo, ok := groupsInfo.(*GroupsInfo)
		if !ok {
			return false
		}

		temp := ansi.Color(thisuid+".keepers:", "green")
		for i, keeper := range thisGroupsInfo.Keepers {
			if i != 0 {
				temp += "_"
			}
			temp += keeper.KID
		}
		res = append(res, temp)
		temp = ansi.Color(thisuid+".providers:", "green")
		temp += strings.Join(thisGroupsInfo.Providers, "_")
		res = append(res, temp)
		return true
	})
	return res, nil
}

func GetProviders() ([]string, error) {
	if !IsKeeperServiceRunning() {
		return nil, ErrKeeperServiceNotReady
	}
	return localPeerInfo.Providers, nil
}

func GetKeepers() ([]string, error) {
	if !IsKeeperServiceRunning() {
		return nil, ErrKeeperServiceNotReady
	}
	return localPeerInfo.Keepers, nil
}

func FlushKeepersAndProviders() error {
	if !IsKeeperServiceRunning() {
		return ErrKeeperServiceNotReady
	}
	return checkConnectedPeer()
}
