package ui

import "strings"

/*
Because every struct can contain many fields and the definition of which fields to show
and which to hide will be verbose, especially when you have a nested struct.
I create a mechanism to make this definition easier for most use cases

In general, the decision of which fields to hide or show is based on three parameters
hide - a list of paths
show - a list of paths
showByDefult - boolean

in case of collision between hide and show list the show is take

each field will be defined dynamically showByDefult for itself and it will be fixed by one of these - (The latter determines)

its parent showByDefult
the user set the filed in the hide list - will be false
the user set the filed in the show list - will be true
if showByDefult is false the field not will show and its children will be set as showByDefult = false unless they are in the hide or show list as mention above

the problem is that the root fields don't have a parent and I need to decide its showByDefult.
so it solved by allowed a user to set it. in case that a user did not set it - it will be calculated by the show list, so if there is at least one field on the root in the show list it will be set to false otherwise is true

*/

type DisplayOpt struct {
	// set the default for the root struct (any root fields will be hidden by default if is true)
	HideAllByDefault bool
	// which field paths to show
	Show []string
	// which field paths to hide
	Hide []string
}


func (opt *DisplayOpt) rootShowByDefault() bool {

	if opt.HideAllByDefault {
		return false
	} else if opt.Show != nil {
		// if there is at least one field on the root of the struct
		for _, path := range opt.Show {
			if !strings.Contains(path, ".") {
				return false
			}
		}
	}
	return true
}

func (opt *DisplayOpt) calcFiledShowByDefult(path []string ,parentShowByDefault bool) bool{
	pathStr := strings.Join(path, ".")

	if opt.Hide != nil {
		if contains(opt.Hide, pathStr) {
			return false
		}
	}
	if opt.Show != nil {
		if contains(opt.Show, pathStr) {
			return true
		}
	}
	return parentShowByDefault
}

