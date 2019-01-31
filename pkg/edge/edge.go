package edge


type Edge map[string]interface{}


func New(uid1, relationship, uid2 string) Edge {
	e := Edge{
		"uid": uid1,
		relationship: map[string]string{
			"uid": uid2,
		},
	}
	return e
}
