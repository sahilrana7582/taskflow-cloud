package project

import "github.com/sahilrana7582/taskflow-cloud/internal/utils"

type CreateRequest struct {
	Key         string  `json:"key"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// UpdateRequest changes only the fields that are present. The key cannot
// change, because job references like ENG-42 are built from it.
type UpdateRequest struct {
	Name        *string                `json:"name"`
	Description utils.Optional[string] `json:"description"`
	Archived    *bool                  `json:"archived"`
}
