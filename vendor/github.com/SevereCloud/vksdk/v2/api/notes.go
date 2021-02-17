package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// NotesAdd creates a new note for the current user.
//
// https://vk.com/dev/notes.add
func (vk *VK) NotesAdd(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("notes.add", &response, params)
	return
}

// NotesCreateComment adds a new comment on a note.
//
// https://vk.com/dev/notes.createComment
func (vk *VK) NotesCreateComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("notes.createComment", &response, params)
	return
}

// NotesDelete deletes a note of the current user.
//
// https://vk.com/dev/notes.delete
func (vk *VK) NotesDelete(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("notes.delete", &response, params)
	return
}

// NotesDeleteComment deletes a comment on a note.
//
// https://vk.com/dev/notes.deleteComment
func (vk *VK) NotesDeleteComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("notes.deleteComment", &response, params)
	return
}

// NotesEdit edits a note of the current user.
//
// https://vk.com/dev/notes.edit
func (vk *VK) NotesEdit(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("notes.edit", &response, params)
	return
}

// NotesEditComment edits a comment on a note.
//
// https://vk.com/dev/notes.editComment
func (vk *VK) NotesEditComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("notes.editComment", &response, params)
	return
}

// NotesGetResponse struct.
type NotesGetResponse struct {
	Count int                `json:"count"`
	Items []object.NotesNote `json:"items"`
}

// NotesGet returns a list of notes created by a user.
//
// https://vk.com/dev/notes.get
func (vk *VK) NotesGet(params Params) (response NotesGetResponse, err error) {
	err = vk.RequestUnmarshal("notes.get", &response, params)
	return
}

// NotesGetByIDResponse struct.
type NotesGetByIDResponse object.NotesNote

// NotesGetByID returns a note by its ID.
//
// https://vk.com/dev/notes.getById
func (vk *VK) NotesGetByID(params Params) (response NotesGetByIDResponse, err error) {
	err = vk.RequestUnmarshal("notes.getById", &response, params)
	return
}

// NotesGetCommentsResponse struct.
type NotesGetCommentsResponse struct {
	Count int                       `json:"count"`
	Items []object.NotesNoteComment `json:"items"`
}

// NotesGetComments returns a list of comments on a note.
//
// https://vk.com/dev/notes.getComments
func (vk *VK) NotesGetComments(params Params) (response NotesGetCommentsResponse, err error) {
	err = vk.RequestUnmarshal("notes.getComments", &response, params)
	return
}

// NotesRestoreComment restores a deleted comment on a note.
//
// https://vk.com/dev/notes.restoreComment
func (vk *VK) NotesRestoreComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("notes.restoreComment", &response, params)
	return
}
