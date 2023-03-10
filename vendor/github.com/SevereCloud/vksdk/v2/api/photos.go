package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// PhotosConfirmTag confirms a tag on a photo.
//
// https://vk.com/dev/photos.confirmTag
func (vk *VK) PhotosConfirmTag(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.confirmTag", &response, params)
	return
}

// PhotosCopy allows to copy a photo to the "Saved photos" album.
//
// https://vk.com/dev/photos.copy
func (vk *VK) PhotosCopy(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.copy", &response, params)
	return
}

// PhotosCreateAlbumResponse struct.
type PhotosCreateAlbumResponse object.PhotosPhotoAlbumFull

// PhotosCreateAlbum creates an empty photo album.
//
// https://vk.com/dev/photos.createAlbum
func (vk *VK) PhotosCreateAlbum(params Params) (response PhotosCreateAlbumResponse, err error) {
	err = vk.RequestUnmarshal("photos.createAlbum", &response, params)
	return
}

// PhotosCreateComment adds a new comment on the photo.
//
// https://vk.com/dev/photos.createComment
func (vk *VK) PhotosCreateComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.createComment", &response, params)
	return
}

// PhotosDelete deletes a photo.
//
// https://vk.com/dev/photos.delete
func (vk *VK) PhotosDelete(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.delete", &response, params)
	return
}

// PhotosDeleteAlbum deletes a photo album belonging to the current user.
//
// https://vk.com/dev/photos.deleteAlbum
func (vk *VK) PhotosDeleteAlbum(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.deleteAlbum", &response, params)
	return
}

// PhotosDeleteComment deletes a comment on the photo.
//
// https://vk.com/dev/photos.deleteComment
func (vk *VK) PhotosDeleteComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.deleteComment", &response, params)
	return
}

// PhotosEdit edits the caption of a photo.
//
// https://vk.com/dev/photos.edit
func (vk *VK) PhotosEdit(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.edit", &response, params)
	return
}

// PhotosEditAlbum edits information about a photo album.
//
// https://vk.com/dev/photos.editAlbum
func (vk *VK) PhotosEditAlbum(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.editAlbum", &response, params)
	return
}

// PhotosEditComment edits a comment on a photo.
//
// https://vk.com/dev/photos.editComment
func (vk *VK) PhotosEditComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.editComment", &response, params)
	return
}

// PhotosGetResponse struct.
type PhotosGetResponse struct {
	Count int                  `json:"count"` // Total number
	Items []object.PhotosPhoto `json:"items"`
}

// PhotosGet returns a list of a user's or community's photos.
//
//	extended=0
//
// https://vk.com/dev/photos.get
func (vk *VK) PhotosGet(params Params) (response PhotosGetResponse, err error) {
	err = vk.RequestUnmarshal("photos.get", &response, params, Params{"extended": false})

	return
}

// PhotosGetExtendedResponse struct.
type PhotosGetExtendedResponse struct {
	Count int                      `json:"count"` // Total number
	Items []object.PhotosPhotoFull `json:"items"`
}

// PhotosGetExtended returns a list of a user's or community's photos.
//
//	extended=1
//
// https://vk.com/dev/photos.get
func (vk *VK) PhotosGetExtended(params Params) (response PhotosGetExtendedResponse, err error) {
	err = vk.RequestUnmarshal("photos.get", &response, params, Params{"extended": true})

	return
}

// PhotosGetAlbumsResponse struct.
type PhotosGetAlbumsResponse struct {
	Count int                           `json:"count"` // Total number
	Items []object.PhotosPhotoAlbumFull `json:"items"`
}

// PhotosGetAlbums returns a list of a user's or community's photo albums.
//
// https://vk.com/dev/photos.getAlbums
func (vk *VK) PhotosGetAlbums(params Params) (response PhotosGetAlbumsResponse, err error) {
	err = vk.RequestUnmarshal("photos.getAlbums", &response, params)
	return
}

// PhotosGetAlbumsCount returns the number of photo albums belonging to a user or community.
//
// https://vk.com/dev/photos.getAlbumsCount
func (vk *VK) PhotosGetAlbumsCount(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.getAlbumsCount", &response, params)
	return
}

// PhotosGetAllResponse struct.
type PhotosGetAllResponse struct {
	Count int                               `json:"count"` // Total number
	Items []object.PhotosPhotoXtrRealOffset `json:"items"`
	More  object.BaseBoolInt                `json:"more"` // Information whether next page is presented
}

// PhotosGetAll returns a list of photos belonging to a user or community, in reverse chronological order.
//
//	extended=0
//
// https://vk.com/dev/photos.getAll
func (vk *VK) PhotosGetAll(params Params) (response PhotosGetAllResponse, err error) {
	err = vk.RequestUnmarshal("photos.getAll", &response, params, Params{"extended": false})

	return
}

// PhotosGetAllExtendedResponse struct.
type PhotosGetAllExtendedResponse struct {
	Count int                                   `json:"count"` // Total number
	Items []object.PhotosPhotoFullXtrRealOffset `json:"items"`
	More  object.BaseBoolInt                    `json:"more"` // Information whether next page is presented
}

// PhotosGetAllExtended returns a list of photos belonging to a user or community, in reverse chronological order.
//
//	extended=1
//
// https://vk.com/dev/photos.getAll
func (vk *VK) PhotosGetAllExtended(params Params) (response PhotosGetAllExtendedResponse, err error) {
	err = vk.RequestUnmarshal("photos.getAll", &response, params, Params{"extended": true})

	return
}

// PhotosGetAllCommentsResponse struct.
type PhotosGetAllCommentsResponse struct {
	Count int                          `json:"count"` // Total number
	Items []object.PhotosCommentXtrPid `json:"items"`
}

// PhotosGetAllComments returns a list of comments on a specific
// photo album or all albums of the user sorted in reverse chronological order.
//
// https://vk.com/dev/photos.getAllComments
func (vk *VK) PhotosGetAllComments(params Params) (response PhotosGetAllCommentsResponse, err error) {
	err = vk.RequestUnmarshal("photos.getAllComments", &response, params)
	return
}

// PhotosGetByIDResponse struct.
type PhotosGetByIDResponse []object.PhotosPhoto

// PhotosGetByID returns information about photos by their IDs.
//
//	extended=0
//
// https://vk.com/dev/photos.getById
func (vk *VK) PhotosGetByID(params Params) (response PhotosGetByIDResponse, err error) {
	err = vk.RequestUnmarshal("photos.getById", &response, params, Params{"extended": false})

	return
}

// PhotosGetByIDExtendedResponse struct.
type PhotosGetByIDExtendedResponse []object.PhotosPhotoFull

// PhotosGetByIDExtended returns information about photos by their IDs.
//
//	extended=1
//
// https://vk.com/dev/photos.getById
func (vk *VK) PhotosGetByIDExtended(params Params) (response PhotosGetByIDExtendedResponse, err error) {
	err = vk.RequestUnmarshal("photos.getById", &response, params, Params{"extended": true})

	return
}

// PhotosGetChatUploadServerResponse struct.
type PhotosGetChatUploadServerResponse struct {
	UploadURL string `json:"upload_url"`
}

// PhotosGetChatUploadServer returns an upload link for chat cover pictures.
//
// https://vk.com/dev/photos.getChatUploadServer
func (vk *VK) PhotosGetChatUploadServer(params Params) (response PhotosGetChatUploadServerResponse, err error) {
	err = vk.RequestUnmarshal("photos.getChatUploadServer", &response, params)
	return
}

// PhotosGetCommentsResponse struct.
type PhotosGetCommentsResponse struct {
	Count      int                      `json:"count"`       // Total number
	RealOffset int                      `json:"real_offset"` // Real offset of the comments
	Items      []object.WallWallComment `json:"items"`
}

// PhotosGetComments returns a list of comments on a photo.
//
//	extended=0
//
// https://vk.com/dev/photos.getComments
func (vk *VK) PhotosGetComments(params Params) (response PhotosGetCommentsResponse, err error) {
	err = vk.RequestUnmarshal("photos.getComments", &response, params, Params{"extended": false})

	return
}

// PhotosGetCommentsExtendedResponse struct.
type PhotosGetCommentsExtendedResponse struct {
	Count      int                      `json:"count"`       // Total number
	RealOffset int                      `json:"real_offset"` // Real offset of the comments
	Items      []object.WallWallComment `json:"items"`
	Profiles   []object.UsersUser       `json:"profiles"`
	Groups     []object.GroupsGroup     `json:"groups"`
}

// PhotosGetCommentsExtended returns a list of comments on a photo.
//
//	extended=1
//
// https://vk.com/dev/photos.getComments
func (vk *VK) PhotosGetCommentsExtended(params Params) (response PhotosGetCommentsExtendedResponse, err error) {
	err = vk.RequestUnmarshal("photos.getComments", &response, params, Params{"extended": true})

	return
}

// PhotosGetMarketAlbumUploadServerResponse struct.
type PhotosGetMarketAlbumUploadServerResponse struct {
	UploadURL string `json:"upload_url"`
}

// PhotosGetMarketAlbumUploadServer returns the server address for market album photo upload.
//
// https://vk.com/dev/photos.getMarketAlbumUploadServer
func (vk *VK) PhotosGetMarketAlbumUploadServer(params Params) (
	response PhotosGetMarketAlbumUploadServerResponse,
	err error,
) {
	err = vk.RequestUnmarshal("photos.getMarketAlbumUploadServer", &response, params)
	return
}

// PhotosGetMarketUploadServerResponse struct.
type PhotosGetMarketUploadServerResponse struct {
	UploadURL string `json:"upload_url"`
}

// PhotosGetMarketUploadServer returns the server address for market photo upload.
//
// https://vk.com/dev/photos.getMarketUploadServer
func (vk *VK) PhotosGetMarketUploadServer(params Params) (response PhotosGetMarketUploadServerResponse, err error) {
	err = vk.RequestUnmarshal("photos.getMarketUploadServer", &response, params)
	return
}

// PhotosGetMessagesUploadServerResponse struct.
type PhotosGetMessagesUploadServerResponse struct {
	AlbumID   int    `json:"album_id"`
	UploadURL string `json:"upload_url"`
	UserID    int    `json:"user_id,omitempty"`
	GroupID   int    `json:"group_id,omitempty"`
}

// PhotosGetMessagesUploadServer returns the server address for photo upload onto a messages.
//
// https://vk.com/dev/photos.getMessagesUploadServer
func (vk *VK) PhotosGetMessagesUploadServer(params Params) (response PhotosGetMessagesUploadServerResponse, err error) {
	err = vk.RequestUnmarshal("photos.getMessagesUploadServer", &response, params)
	return
}

// PhotosGetNewTagsResponse struct.
type PhotosGetNewTagsResponse struct {
	Count int                            `json:"count"` // Total number
	Items []object.PhotosPhotoXtrTagInfo `json:"items"`
}

// PhotosGetNewTags returns a list of photos with tags that have not been viewed.
//
// https://vk.com/dev/photos.getNewTags
func (vk *VK) PhotosGetNewTags(params Params) (response PhotosGetNewTagsResponse, err error) {
	err = vk.RequestUnmarshal("photos.getNewTags", &response, params)
	return
}

// PhotosGetOwnerCoverPhotoUploadServerResponse struct.
type PhotosGetOwnerCoverPhotoUploadServerResponse struct {
	UploadURL string `json:"upload_url"`
}

// PhotosGetOwnerCoverPhotoUploadServer receives server address for uploading community cover.
//
// https://vk.com/dev/photos.getOwnerCoverPhotoUploadServer
func (vk *VK) PhotosGetOwnerCoverPhotoUploadServer(params Params) (
	response PhotosGetOwnerCoverPhotoUploadServerResponse,
	err error,
) {
	err = vk.RequestUnmarshal("photos.getOwnerCoverPhotoUploadServer", &response, params)
	return
}

// PhotosGetOwnerPhotoUploadServerResponse struct.
type PhotosGetOwnerPhotoUploadServerResponse struct {
	UploadURL string `json:"upload_url"`
}

// PhotosGetOwnerPhotoUploadServer returns an upload server address for a
// profile or community photo.
//
// https://vk.com/dev/photos.getOwnerPhotoUploadServer
func (vk *VK) PhotosGetOwnerPhotoUploadServer(params Params) (
	response PhotosGetOwnerPhotoUploadServerResponse,
	err error,
) {
	err = vk.RequestUnmarshal("photos.getOwnerPhotoUploadServer", &response, params)
	return
}

// PhotosGetTagsResponse struct.
type PhotosGetTagsResponse []object.PhotosPhotoTag

// PhotosGetTags returns a list of tags on a photo.
//
// https://vk.com/dev/photos.getTags
func (vk *VK) PhotosGetTags(params Params) (response PhotosGetTagsResponse, err error) {
	err = vk.RequestUnmarshal("photos.getTags", &response, params)
	return
}

// PhotosGetUploadServerResponse struct.
type PhotosGetUploadServerResponse object.PhotosPhotoUpload

// PhotosGetUploadServer returns the server address for photo upload.
//
// https://vk.com/dev/photos.getUploadServer
func (vk *VK) PhotosGetUploadServer(params Params) (response PhotosGetUploadServerResponse, err error) {
	err = vk.RequestUnmarshal("photos.getUploadServer", &response, params)
	return
}

// PhotosGetUserPhotosResponse struct.
type PhotosGetUserPhotosResponse struct {
	Count int                  `json:"count"` // Total number
	Items []object.PhotosPhoto `json:"items"`
}

// PhotosGetUserPhotos returns a list of photos in which a user is tagged.
//
//	extended=0
//
// https://vk.com/dev/photos.getUserPhotos
func (vk *VK) PhotosGetUserPhotos(params Params) (response PhotosGetUserPhotosResponse, err error) {
	err = vk.RequestUnmarshal("photos.getUserPhotos", &response, params, Params{"extended": false})

	return
}

// PhotosGetUserPhotosExtendedResponse struct.
type PhotosGetUserPhotosExtendedResponse struct {
	Count int                      `json:"count"` // Total number
	Items []object.PhotosPhotoFull `json:"items"`
}

// PhotosGetUserPhotosExtended returns a list of photos in which a user is tagged.
//
//	extended=1
//
// https://vk.com/dev/photos.getUserPhotos
func (vk *VK) PhotosGetUserPhotosExtended(params Params) (response PhotosGetUserPhotosExtendedResponse, err error) {
	err = vk.RequestUnmarshal("photos.getUserPhotos", &response, params, Params{"extended": true})

	return
}

// PhotosGetWallUploadServerResponse struct.
type PhotosGetWallUploadServerResponse object.PhotosPhotoUpload

// PhotosGetWallUploadServer returns the server address for photo upload onto a user's wall.
//
// https://vk.com/dev/photos.getWallUploadServer
func (vk *VK) PhotosGetWallUploadServer(params Params) (response PhotosGetWallUploadServerResponse, err error) {
	err = vk.RequestUnmarshal("photos.getWallUploadServer", &response, params)
	return
}

// PhotosMakeCover makes a photo into an album cover.
//
// https://vk.com/dev/photos.makeCover
func (vk *VK) PhotosMakeCover(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.makeCover", &response, params)
	return
}

// PhotosMove a photo from one album to another.
//
// https://vk.com/dev/photos.moveMoves
func (vk *VK) PhotosMove(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.move", &response, params)
	return
}

// PhotosPutTag adds a tag on the photo.
//
// https://vk.com/dev/photos.putTag
func (vk *VK) PhotosPutTag(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.putTag", &response, params)
	return
}

// PhotosRemoveTag removes a tag from a photo.
//
// https://vk.com/dev/photos.removeTag
func (vk *VK) PhotosRemoveTag(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.removeTag", &response, params)
	return
}

// PhotosReorderAlbums reorders the album in the list of user albums.
//
// https://vk.com/dev/photos.reorderAlbums
func (vk *VK) PhotosReorderAlbums(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.reorderAlbums", &response, params)
	return
}

// PhotosReorderPhotos reorders the photo in the list of photos of the user album.
//
// https://vk.com/dev/photos.reorderPhotos
func (vk *VK) PhotosReorderPhotos(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.reorderPhotos", &response, params)
	return
}

// PhotosReport reports (submits a complaint about) a photo.
//
// https://vk.com/dev/photos.report
func (vk *VK) PhotosReport(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.report", &response, params)
	return
}

// PhotosReportComment reports (submits a complaint about) a comment on a photo.
//
// https://vk.com/dev/photos.reportComment
func (vk *VK) PhotosReportComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.reportComment", &response, params)
	return
}

// PhotosRestore restores a deleted photo.
//
// https://vk.com/dev/photos.restore
func (vk *VK) PhotosRestore(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.restore", &response, params)
	return
}

// PhotosRestoreComment restores a deleted comment on a photo.
//
// https://vk.com/dev/photos.restoreComment
func (vk *VK) PhotosRestoreComment(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("photos.restoreComment", &response, params)
	return
}

// PhotosSaveResponse struct.
type PhotosSaveResponse []object.PhotosPhoto

// PhotosSave saves photos after successful uploading.
//
// https://vk.com/dev/photos.save
func (vk *VK) PhotosSave(params Params) (response PhotosSaveResponse, err error) {
	err = vk.RequestUnmarshal("photos.save", &response, params)
	return
}

// PhotosSaveMarketAlbumPhotoResponse struct.
type PhotosSaveMarketAlbumPhotoResponse []object.PhotosPhoto

// PhotosSaveMarketAlbumPhoto photo Saves market album photos after successful uploading.
//
// https://vk.com/dev/photos.saveMarketAlbumPhoto
func (vk *VK) PhotosSaveMarketAlbumPhoto(params Params) (response PhotosSaveMarketAlbumPhotoResponse, err error) {
	err = vk.RequestUnmarshal("photos.saveMarketAlbumPhoto", &response, params)
	return
}

// PhotosSaveMarketPhotoResponse struct.
type PhotosSaveMarketPhotoResponse []object.PhotosPhoto

// PhotosSaveMarketPhoto saves market photos after successful uploading.
//
// https://vk.com/dev/photos.saveMarketPhoto
func (vk *VK) PhotosSaveMarketPhoto(params Params) (response PhotosSaveMarketPhotoResponse, err error) {
	err = vk.RequestUnmarshal("photos.saveMarketPhoto", &response, params)
	return
}

// PhotosSaveMessagesPhotoResponse struct.
type PhotosSaveMessagesPhotoResponse []object.PhotosPhoto

// PhotosSaveMessagesPhoto saves a photo after being successfully.
//
// https://vk.com/dev/photos.saveMessagesPhoto
func (vk *VK) PhotosSaveMessagesPhoto(params Params) (response PhotosSaveMessagesPhotoResponse, err error) {
	err = vk.RequestUnmarshal("photos.saveMessagesPhoto", &response, params)
	return
}

// PhotosSaveOwnerCoverPhotoResponse struct.
type PhotosSaveOwnerCoverPhotoResponse struct {
	Images []object.PhotosImage `json:"images"`
}

// PhotosSaveOwnerCoverPhoto saves cover photo after successful uploading.
//
// https://vk.com/dev/photos.saveOwnerCoverPhoto
func (vk *VK) PhotosSaveOwnerCoverPhoto(params Params) (response PhotosSaveOwnerCoverPhotoResponse, err error) {
	err = vk.RequestUnmarshal("photos.saveOwnerCoverPhoto", &response, params)
	return
}

// PhotosSaveOwnerPhotoResponse struct.
type PhotosSaveOwnerPhotoResponse struct {
	PhotoHash string `json:"photo_hash"`
	// BUG(VK): returns false
	// PhotoSrc      string `json:"photo_src"`
	// PhotoSrcBig   string `json:"photo_src_big"`
	// PhotoSrcSmall string `json:"photo_src_small"`
	Saved  int `json:"saved"`
	PostID int `json:"post_id"`
}

// PhotosSaveOwnerPhoto saves a profile or community photo.
//
// https://vk.com/dev/photos.saveOwnerPhoto
func (vk *VK) PhotosSaveOwnerPhoto(params Params) (response PhotosSaveOwnerPhotoResponse, err error) {
	err = vk.RequestUnmarshal("photos.saveOwnerPhoto", &response, params)
	return
}

// PhotosSaveWallPhotoResponse struct.
type PhotosSaveWallPhotoResponse []object.PhotosPhoto

// PhotosSaveWallPhoto saves a photo to a user's or community's wall after being uploaded.
//
// https://vk.com/dev/photos.saveWallPhoto
func (vk *VK) PhotosSaveWallPhoto(params Params) (response PhotosSaveWallPhotoResponse, err error) {
	err = vk.RequestUnmarshal("photos.saveWallPhoto", &response, params)
	return
}

// PhotosSearchResponse struct.
type PhotosSearchResponse struct {
	Count int                      `json:"count"` // Total number
	Items []object.PhotosPhotoFull `json:"items"`
}

// PhotosSearch returns a list of photos.
//
// https://vk.com/dev/photos.search
func (vk *VK) PhotosSearch(params Params) (response PhotosSearchResponse, err error) {
	err = vk.RequestUnmarshal("photos.search", &response, params)
	return
}
