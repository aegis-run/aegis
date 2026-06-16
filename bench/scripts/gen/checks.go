package gen

type ResourceRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type Check struct {
	Resource   ResourceRef `json:"resource"`
	Permission string      `json:"permission"`
	Actor      ResourceRef `json:"actor"`
	Expected   bool        `json:"expected"`
	Class      string      `json:"class"`
}
