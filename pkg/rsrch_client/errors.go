package rsrch_client

//   WAIT_FOR_OFER we intend to take this struct from researcher-ui repository, still working on it
type Error struct {
	Message    	string `json:"message"`
	Details  	string `json:"details"`
	Status 		int    `json:"status"`
}
