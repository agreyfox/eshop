package user

// Permissions describe a user's permissions.
type Permissions struct {
	Admin    bool `json:"admin"`
	Execute  bool `json:"execute"`
	Create   bool `json:"create"`
	Modify   bool `json:"modify"`
	Delete   bool `json:"delete"`
	Share    bool `json:"share"`
	Download bool `json:"download"`
}

var (
	AdminPermmission = Permissions{
		Admin:    true,
		Execute:  true,
		Create:   true,
		Modify:   true,
		Delete:   true,
		Share:    true,
		Download: true,
	}
	CustomerPermission = Permissions{
		Admin:    false,
		Execute:  true,
		Create:   true,
		Modify:   true,
		Delete:   false,
		Share:    true,
		Download: true,
	}
)
