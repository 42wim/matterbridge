package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"

	"github.com/SevereCloud/vksdk/v2/object"
)

// UploadFile uploading file.
func (vk *VK) UploadFile(url string, file io.Reader, fieldname, filename string) (bodyContent []byte, err error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldname, filename)
	if err != nil {
		return
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return
	}

	contentType := writer.FormDataContentType()
	_ = writer.Close()

	resp, err := vk.Client.Post(url, contentType, body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	bodyContent, err = io.ReadAll(resp.Body)

	return
}

// uploadPhoto uploading Photos into Album.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: width+height not more than 14000 px, file size up to 50 Mb,
// aspect ratio of at least 1:20.
func (vk *VK) uploadPhoto(params Params, file io.Reader) (response PhotosSaveResponse, err error) {
	uploadServer, err := vk.PhotosGetUploadServer(params)
	if err != nil {
		return
	}

	bodyContent, err := vk.UploadFile(uploadServer.UploadURL, file, "file1", "file1.jpeg")
	if err != nil {
		return
	}

	var handler object.PhotosPhotoUploadResponse

	err = json.Unmarshal(bodyContent, &handler)
	if err != nil {
		return
	}

	response, err = vk.PhotosSave(Params{
		"server":      handler.Server,
		"photos_list": handler.PhotosList,
		"aid":         handler.AID,
		"hash":        handler.Hash,
		"album_id":    params["album_id"],
		"group_id":    params["group_id"],
	})

	return
}

// UploadPhoto uploading Photos into User Album.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: width+height not more than 14000 px, file size up to 50 Mb,
// aspect ratio of at least 1:20.
func (vk *VK) UploadPhoto(albumID int, file io.Reader) (response PhotosSaveResponse, err error) {
	response, err = vk.uploadPhoto(Params{
		"album_id": albumID,
	}, file)

	return
}

// UploadPhotoGroup uploading Photos into Group Album.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: width+height not more than 14000 px, file size up to 50 Mb,
// aspect ratio of at least 1:20.
func (vk *VK) UploadPhotoGroup(groupID, albumID int, file io.Reader) (response PhotosSaveResponse, err error) {
	response, err = vk.uploadPhoto(Params{
		"album_id": albumID,
		"group_id": groupID,
	}, file)

	return
}

// uploadWallPhoto uploading Photos on Wall.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: width+height not more than 14000 px, file size up to 50 Mb,
// aspect ratio of at least 1:20.
func (vk *VK) uploadWallPhoto(params Params, file io.Reader) (response PhotosSaveWallPhotoResponse, err error) {
	uploadServer, err := vk.PhotosGetWallUploadServer(params)
	if err != nil {
		return
	}

	bodyContent, err := vk.UploadFile(uploadServer.UploadURL, file, "photo", "photo.jpeg")
	if err != nil {
		return
	}

	var handler object.PhotosWallUploadResponse

	err = json.Unmarshal(bodyContent, &handler)
	if err != nil {
		return
	}

	response, err = vk.PhotosSaveWallPhoto(Params{
		"server":   handler.Server,
		"photo":    handler.Photo,
		"hash":     handler.Hash,
		"group_id": params["group_id"],
	})

	return
}

// UploadWallPhoto uploading Photos on User Wall.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: width+height not more than 14000 px, file size up to 50 Mb,
// aspect ratio of at least 1:20.
func (vk *VK) UploadWallPhoto(file io.Reader) (response PhotosSaveWallPhotoResponse, err error) {
	response, err = vk.uploadWallPhoto(Params{}, file)
	return
}

// UploadGroupWallPhoto uploading Photos on Group Wall.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: width+height not more than 14000 px, file size up to 50 Mb,
// aspect ratio of at least 1:20.
func (vk *VK) UploadGroupWallPhoto(groupID int, file io.Reader) (response PhotosSaveWallPhotoResponse, err error) {
	response, err = vk.uploadWallPhoto(Params{
		"group_id": groupID,
	}, file)

	return
}

// uploadOwnerPhoto uploading Photos into User Profile or Community
// To upload a photo to a community send its negative id in the owner_id parameter.
//
// Following parameters can be sent in addition:
// squareCrop in x,y,w (no quotes) format where x and y are the coordinates of
// the preview upper-right corner and w is square side length.
// That will create a square preview for a photo.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: size not less than 200x200px, aspect ratio from 0.25 to 3,
// width+height not more than 14000 px, file size up to 50 Mb.
func (vk *VK) uploadOwnerPhoto(params Params, squareCrop string, file io.Reader) (
	response PhotosSaveOwnerPhotoResponse,
	err error,
) {
	uploadServer, err := vk.PhotosGetOwnerPhotoUploadServer(params)
	if err != nil {
		return
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("photo", "photo.jpeg")
	if err != nil {
		return
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return
	}

	contentType := writer.FormDataContentType()

	if squareCrop != "" {
		err = writer.WriteField("_square_crop", squareCrop)
		if err != nil {
			return
		}
	}

	_ = writer.Close()

	resp, err := vk.Client.Post(uploadServer.UploadURL, contentType, body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	bodyContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var handler object.PhotosOwnerUploadResponse

	err = json.Unmarshal(bodyContent, &handler)
	if err != nil {
		return
	}

	response, err = vk.PhotosSaveOwnerPhoto(Params{
		"server": handler.Server,
		"photo":  handler.Photo,
		"hash":   handler.Hash,
	})

	return response, err
}

// UploadUserPhoto uploading Photos into User Profile.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: size not less than 200x200px, aspect ratio from 0.25 to 3,
// width+height not more than 14000 px, file size up to 50 Mb.
func (vk *VK) UploadUserPhoto(file io.Reader) (response PhotosSaveOwnerPhotoResponse, err error) {
	response, err = vk.uploadOwnerPhoto(Params{}, "", file)
	return
}

// UploadOwnerPhoto uploading Photos into User Profile or Community
// To upload a photo to a community send its negative id in the owner_id parameter.
//
// Following parameters can be sent in addition:
// squareCrop in x,y,w (no quotes) format where x and y are the coordinates of
// the preview upper-right corner and w is square side length.
// That will create a square preview for a photo.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: size not less than 200x200px, aspect ratio from 0.25 to 3,
// width+height not more than 14000 px, file size up to 50 Mb.
func (vk *VK) UploadOwnerPhoto(ownerID int, squareCrop string, file io.Reader) (
	response PhotosSaveOwnerPhotoResponse,
	err error,
) {
	response, err = vk.uploadOwnerPhoto(Params{
		"owner_id": ownerID,
	}, squareCrop, file)

	return
}

// UploadMessagesPhoto uploading Photos into a Private Message.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: width+height not more than 14000 px, file size up to 50 Mb,
// aspect ratio of at least 1:20.
func (vk *VK) UploadMessagesPhoto(peerID int, file io.Reader) (response PhotosSaveMessagesPhotoResponse, err error) {
	uploadServer, err := vk.PhotosGetMessagesUploadServer(Params{
		"peer_id": peerID,
	})
	if err != nil {
		return
	}

	bodyContent, err := vk.UploadFile(uploadServer.UploadURL, file, "photo", "photo.jpeg")
	if err != nil {
		return
	}

	var handler object.PhotosMessageUploadResponse

	err = json.Unmarshal(bodyContent, &handler)
	if err != nil {
		return
	}

	response, err = vk.PhotosSaveMessagesPhoto(Params{
		"server": handler.Server,
		"photo":  handler.Photo,
		"hash":   handler.Hash,
	})

	return
}

// uploadChatPhoto uploading a Main Photo to a Group Chat.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: size not less than 200x200px,
// width+height not more than 14000 px, file size up to 50 Mb,
// aspect ratio of at least 1:20.
func (vk *VK) uploadChatPhoto(params Params, file io.Reader) (response MessagesSetChatPhotoResponse, err error) {
	uploadServer, err := vk.PhotosGetChatUploadServer(params)
	if err != nil {
		return
	}

	bodyContent, err := vk.UploadFile(uploadServer.UploadURL, file, "file", "photo.jpeg")
	if err != nil {
		return
	}

	var handler object.PhotosChatUploadResponse

	err = json.Unmarshal(bodyContent, &handler)
	if err != nil {
		return
	}

	response, err = vk.MessagesSetChatPhoto(Params{
		"file": handler.Response,
	})

	return
}

// UploadChatPhoto uploading a Main Photo to a Group Chat without crop.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: size not less than 200x200px,
// width+height not more than 14000 px, file size up to 50 Mb,
// aspect ratio of at least 1:20.
func (vk *VK) UploadChatPhoto(chatID int, file io.Reader) (response MessagesSetChatPhotoResponse, err error) {
	response, err = vk.uploadChatPhoto(Params{
		"chat_id": chatID,
	}, file)

	return
}

// UploadChatPhotoCrop uploading a Main Photo to a Group Chat with crop.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: size not less than 200x200px,
// width+height not more than 14000 px, file size up to 50 Mb,
// aspect ratio of at least 1:20.
func (vk *VK) UploadChatPhotoCrop(chatID, cropX, cropY, cropWidth int, file io.Reader) (
	response MessagesSetChatPhotoResponse,
	err error,
) {
	response, err = vk.uploadChatPhoto(Params{
		"chat_id":    chatID,
		"crop_x":     cropX,
		"crop_y":     cropY,
		"crop_width": cropWidth,
	}, file)

	return
}

// uploadMarketPhoto uploading a Market Item Photo.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: size not less than 400x400px,
// width+height not more than 14000 px, file size up to 50 Mb,
// aspect ratio of at least 1:20.
func (vk *VK) uploadMarketPhoto(params Params, file io.Reader) (response PhotosSaveMarketPhotoResponse, err error) {
	uploadServer, err := vk.PhotosGetMarketUploadServer(params)
	if err != nil {
		return
	}

	bodyContent, err := vk.UploadFile(uploadServer.UploadURL, file, "file", "photo.jpeg")
	if err != nil {
		return
	}

	var handler object.PhotosMarketUploadResponse

	err = json.Unmarshal(bodyContent, &handler)
	if err != nil {
		return
	}

	response, err = vk.PhotosSaveMarketPhoto(Params{
		"group_id":  params["group_id"],
		"server":    handler.Server,
		"photo":     handler.Photo,
		"hash":      handler.Hash,
		"crop_data": handler.CropData,
		"crop_hash": handler.CropHash,
	})

	return
}

// UploadMarketPhoto uploading a Market Item Photo without crop.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: size not less than 400x400px,
// width+height not more than 14000 px, file size up to 50 Mb,
// aspect ratio of at least 1:20.
func (vk *VK) UploadMarketPhoto(groupID int, mainPhoto bool, file io.Reader) (
	response PhotosSaveMarketPhotoResponse,
	err error,
) {
	response, err = vk.uploadMarketPhoto(Params{
		"group_id":   groupID,
		"main_photo": mainPhoto,
	}, file)

	return
}

// UploadMarketPhotoCrop uploading a Market Item Photo with crop.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: size not less than 400x400px,
// width+height not more than 14000 px, file size up to 50 Mb,
// aspect ratio of at least 1:20.
func (vk *VK) UploadMarketPhotoCrop(groupID, cropX, cropY, cropWidth int, file io.Reader) (
	response PhotosSaveMarketPhotoResponse,
	err error,
) {
	response, err = vk.uploadMarketPhoto(Params{
		"group_id":   groupID,
		"main_photo": true,
		"crop_x":     cropX,
		"crop_y":     cropY,
		"crop_width": cropWidth,
	}, file)

	return
}

// UploadMarketAlbumPhoto uploading a Main Photo to a Group Chat.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: size not less than 1280x720px,
// width+height not more than 14000 px, file size up to 50 Mb,
// aspect ratio of at least 1:20.
func (vk *VK) UploadMarketAlbumPhoto(groupID int, file io.Reader) (
	response PhotosSaveMarketAlbumPhotoResponse,
	err error,
) {
	uploadServer, err := vk.PhotosGetMarketAlbumUploadServer(Params{
		"group_id": groupID,
	})
	if err != nil {
		return
	}

	bodyContent, err := vk.UploadFile(uploadServer.UploadURL, file, "file", "photo.jpeg")
	if err != nil {
		return
	}

	var handler object.PhotosMarketAlbumUploadResponse

	err = json.Unmarshal(bodyContent, &handler)
	if err != nil {
		return
	}

	return vk.PhotosSaveMarketAlbumPhoto(Params{
		"group_id": groupID,
		"server":   handler.Server,
		"photo":    handler.Photo,
		"hash":     handler.Hash,
	})
}

// UploadVideo uploading Video Files.
//
// Supported formats: AVI, MP4, 3GP, MPEG, MOV, FLV, WMV.
func (vk *VK) UploadVideo(params Params, file io.Reader) (response VideoSaveResponse, err error) {
	response, err = vk.VideoSave(params)
	if err != nil {
		return
	}

	bodyContent, err := vk.UploadFile(response.UploadURL, file, "video_file", "video.mp4")
	if err != nil {
		return
	}

	var videoUploadError UploadError

	err = json.Unmarshal(bodyContent, &videoUploadError)
	if err != nil {
		return
	}

	if videoUploadError.Code != 0 {
		err = &videoUploadError
	}

	return
}

// uploadDoc uploading Documents.
//
// Supported formats: any formats excepting mp3 and executable files.
//
// Limits: file size up to 200 MB.
func (vk *VK) uploadDoc(url, title, tags string, file io.Reader) (response DocsSaveResponse, err error) {
	bodyContent, err := vk.UploadFile(url, file, "file", title)
	if err != nil {
		return
	}

	var docUploadError UploadError

	err = json.Unmarshal(bodyContent, &docUploadError)
	if err != nil {
		return
	}

	if docUploadError.Err != "" {
		err = &docUploadError
		return
	}

	var handler object.DocsDocUploadResponse

	err = json.Unmarshal(bodyContent, &handler)
	if err != nil {
		return
	}

	response, err = vk.DocsSave(Params{
		"file":  handler.File,
		"title": title,
		"tags":  tags,
	})

	return response, err
}

// UploadDoc uploading Documents.
//
// Supported formats: any formats excepting mp3 and executable files.
//
// Limits: file size up to 200 MB.
func (vk *VK) UploadDoc(title, tags string, file io.Reader) (response DocsSaveResponse, err error) {
	uploadServer, err := vk.DocsGetUploadServer(nil)
	if err != nil {
		return
	}

	response, err = vk.uploadDoc(uploadServer.UploadURL, title, tags, file)

	return
}

// UploadGroupDoc uploading Documents into Community.
//
// Supported formats: any formats excepting mp3 and executable files.
//
// Limits: file size up to 200 MB.
func (vk *VK) UploadGroupDoc(groupID int, title, tags string, file io.Reader) (response DocsSaveResponse, err error) {
	uploadServer, err := vk.DocsGetUploadServer(Params{
		"group_id": groupID,
	})
	if err != nil {
		return
	}

	response, err = vk.uploadDoc(uploadServer.UploadURL, title, tags, file)

	return
}

// UploadWallDoc uploading Documents on Wall.
//
// Supported formats: any formats excepting mp3 and executable files.
//
// Limits: file size up to 200 MB.
func (vk *VK) UploadWallDoc(title, tags string, file io.Reader) (response DocsSaveResponse, err error) {
	uploadServer, err := vk.DocsGetWallUploadServer(nil)
	if err != nil {
		return
	}

	response, err = vk.uploadDoc(uploadServer.UploadURL, title, tags, file)

	return
}

// UploadGroupWallDoc uploading Documents on Group Wall.
//
// Supported formats: any formats excepting mp3 and executable files.
//
// Limits: file size up to 200 MB.
func (vk *VK) UploadGroupWallDoc(groupID int, title, tags string, file io.Reader) (
	response DocsSaveResponse,
	err error,
) {
	uploadServer, err := vk.DocsGetWallUploadServer(Params{
		"group_id": groupID,
	})
	if err != nil {
		return
	}

	response, err = vk.uploadDoc(uploadServer.UploadURL, title, tags, file)

	return
}

// UploadMessagesDoc uploading Documents into a Private Message.
//
// Supported formats: any formats excepting mp3 and executable files.
//
// Limits: file size up to 200 MB.
func (vk *VK) UploadMessagesDoc(peerID int, typeDoc, title, tags string, file io.Reader) (
	response DocsSaveResponse,
	err error,
) {
	uploadServer, err := vk.DocsGetMessagesUploadServer(Params{
		"peer_id": peerID,
		"type":    typeDoc,
	})
	if err != nil {
		return
	}

	response, err = vk.uploadDoc(uploadServer.UploadURL, title, tags, file)

	return
}

// UploadOwnerCoverPhoto uploading a Main Photo to a Group Chat.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: minimum photo size 795x200px, width+height not more than 14000px,
// file size up to 50 MB. Recommended size: 1590x400px.
func (vk *VK) UploadOwnerCoverPhoto(groupID, cropX, cropY, cropX2, cropY2 int, file io.Reader) (
	response PhotosSaveOwnerCoverPhotoResponse,
	err error,
) {
	uploadServer, err := vk.PhotosGetOwnerCoverPhotoUploadServer(Params{
		"group_id": groupID,
		"crop_x":   cropX,
		"crop_y":   cropY,
		"crop_x2":  cropX2,
		"crop_y2":  cropY2,
	})
	if err != nil {
		return
	}

	bodyContent, err := vk.UploadFile(uploadServer.UploadURL, file, "photo", "photo.jpeg")
	if err != nil {
		return
	}

	var handler object.PhotosOwnerUploadResponse

	err = json.Unmarshal(bodyContent, &handler)
	if err != nil {
		return
	}

	return vk.PhotosSaveOwnerCoverPhoto(Params{
		"photo": handler.Photo,
		"hash":  handler.Hash,
	})
}

// UploadStories struct.
type UploadStories struct {
	UploadResult string `json:"upload_result"`
	Sig          string `json:"_sig"`
}

type rawUploadStoriesPhoto struct {
	Response UploadStories `json:"response"`
	Error    struct {
		ErrorCode int    `json:"error_code"`
		Type      string `json:"type"`
	} `json:"error"`
}

type rawUploadStoriesVideo struct {
	Response UploadStories `json:"response"`
	UploadError
}

// UploadStoriesPhoto uploading Story.
//
// Supported formats: JPG, PNG, GIF.
// Limits: sum of with and height no more than 14000px, file size no
// more than 10 MB. Video format: h264 video, aac audio,
// maximum 720х1280, 30fps.
//
// https://vk.com/dev/stories.getPhotoUploadServer
func (vk *VK) UploadStoriesPhoto(params Params, file io.Reader) (response StoriesSaveResponse, err error) {
	uploadServer, err := vk.StoriesGetPhotoUploadServer(params)
	if err != nil {
		return
	}

	bodyContent, err := vk.UploadFile(uploadServer.UploadURL, file, "file", "file.jpeg")
	if err != nil {
		return
	}

	var handler rawUploadStoriesPhoto

	err = json.Unmarshal(bodyContent, &handler)
	if err != nil {
		return
	}

	if handler.Error.ErrorCode != 0 {
		err = &UploadError{
			Code: handler.Error.ErrorCode,
			Err:  handler.Error.Type,
		}

		return response, err
	}

	response, err = vk.StoriesSave(Params{
		"upload_results": handler.Response.UploadResult,
	})

	return response, err
}

// UploadStoriesVideo uploading Story.
//
// Video format: h264 video, aac audio, maximum 720х1280, 30fps.
func (vk *VK) UploadStoriesVideo(params Params, file io.Reader) (response StoriesSaveResponse, err error) {
	uploadServer, err := vk.StoriesGetVideoUploadServer(params)
	if err != nil {
		return
	}

	bodyContent, err := vk.UploadFile(uploadServer.UploadURL, file, "video_file", "video.mp4")
	if err != nil {
		return
	}

	var handler rawUploadStoriesVideo

	err = json.Unmarshal(bodyContent, &handler)
	if err != nil {
		return
	}

	if handler.UploadError.Code != 0 {
		return response, &handler.UploadError
	}

	response, err = vk.StoriesSave(Params{
		"upload_results": handler.Response.UploadResult,
	})

	return response, err
}

// uploadPollsPhoto uploading a Poll Photo.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: minimum photo size 795x200px, width+height not more than 14000px,
// file size up to 50 MB. Recommended size: 1590x400px.
func (vk *VK) uploadPollsPhoto(params Params, file io.Reader) (response PollsSavePhotoResponse, err error) {
	uploadServer, err := vk.PollsGetPhotoUploadServer(params)
	if err != nil {
		return
	}

	bodyContent, err := vk.UploadFile(uploadServer.UploadURL, file, "photo", "photo.jpeg")
	if err != nil {
		return
	}

	var handler object.PollsPhotoUploadResponse

	err = json.Unmarshal(bodyContent, &handler)
	if err != nil {
		return
	}

	response, err = vk.PollsSavePhoto(Params{
		"photo": handler.Photo,
		"hash":  handler.Hash,
	})

	return
}

// UploadPollsPhoto uploading a Poll Photo.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: minimum photo size 795x200px, width+height not more than 14000px,
// file size up to 50 MB. Recommended size: 1590x400px.
func (vk *VK) UploadPollsPhoto(file io.Reader) (response PollsSavePhotoResponse, err error) {
	return vk.uploadPollsPhoto(Params{}, file)
}

// UploadOwnerPollsPhoto uploading a Poll Photo.
//
// Supported formats: JPG, PNG, GIF.
//
// Limits: minimum photo size 795x200px, width+height not more than 14000px,
// file size up to 50 MB. Recommended size: 1590x400px.
func (vk *VK) UploadOwnerPollsPhoto(ownerID int, file io.Reader) (response PollsSavePhotoResponse, err error) {
	return vk.uploadPollsPhoto(Params{"owner_id": ownerID}, file)
}

type uploadPrettyCardsPhotoHandler struct {
	Photo   string `json:"photo"`
	ErrCode int    `json:"errcode"`
}

// UploadPrettyCardsPhoto uploading a Pretty Card Photo.
//
// Supported formats: JPG, PNG, GIF.
func (vk *VK) UploadPrettyCardsPhoto(file io.Reader) (response string, err error) {
	uploadURL, err := vk.PrettyCardsGetUploadURL(nil)
	if err != nil {
		return
	}

	bodyContent, err := vk.UploadFile(uploadURL, file, "file", "photo.jpg")
	if err != nil {
		return
	}

	var handler uploadPrettyCardsPhotoHandler

	err = json.Unmarshal(bodyContent, &handler)
	if err != nil {
		return
	}

	response = handler.Photo

	if handler.ErrCode != 0 {
		err = &UploadError{Code: handler.ErrCode}
	}

	return
}

type uploadLeadFormsPhotoHandler struct {
	Photo   string `json:"photo"`
	ErrCode int    `json:"errcode"`
}

// UploadLeadFormsPhoto uploading a Pretty Card Photo.
//
// Supported formats: JPG, PNG, GIF.
func (vk *VK) UploadLeadFormsPhoto(file io.Reader) (response string, err error) {
	uploadURL, err := vk.LeadFormsGetUploadURL(nil)
	if err != nil {
		return
	}

	bodyContent, err := vk.UploadFile(uploadURL, file, "file", "photo.jpg")
	if err != nil {
		return
	}

	var handler uploadLeadFormsPhotoHandler

	err = json.Unmarshal(bodyContent, &handler)
	if err != nil {
		return
	}

	response = handler.Photo

	if handler.ErrCode != 0 {
		err = &UploadError{Code: handler.ErrCode}
	}

	return
}

// UploadAppImage uploading a Image into App collection for community app widgets.
func (vk *VK) UploadAppImage(imageType string, file io.Reader) (response object.AppWidgetsImage, err error) {
	uploadServer, err := vk.AppWidgetsGetAppImageUploadServer(Params{
		"image_type": imageType,
	})
	if err != nil {
		return
	}

	bodyContent, err := vk.UploadFile(uploadServer.UploadURL, file, "image", "photo.jpeg")
	if err != nil {
		return
	}

	var handler object.AppWidgetsAppImageUploadResponse

	err = json.Unmarshal(bodyContent, &handler)
	if err != nil {
		return
	}

	response, err = vk.AppWidgetsSaveAppImage(Params{
		"image": handler.Image,
		"hash":  handler.Hash,
	})

	return
}

// UploadGroupImage uploading a Image into Community collection for community app widgets.
func (vk *VK) UploadGroupImage(imageType string, file io.Reader) (response object.AppWidgetsImage, err error) {
	uploadServer, err := vk.AppWidgetsGetGroupImageUploadServer(Params{
		"image_type": imageType,
	})
	if err != nil {
		return
	}

	bodyContent, err := vk.UploadFile(uploadServer.UploadURL, file, "image", "photo.jpeg")
	if err != nil {
		return
	}

	var handler object.AppWidgetsGroupImageUploadResponse

	err = json.Unmarshal(bodyContent, &handler)
	if err != nil {
		return
	}

	response, err = vk.AppWidgetsSaveGroupImage(Params{
		"image": handler.Image,
		"hash":  handler.Hash,
	})

	return
}

// UploadMarusiaPicture uploading picture.
//
// Limits: height not more than 600 px,
// aspect ratio of at least 2:1.
func (vk *VK) UploadMarusiaPicture(file io.Reader) (response MarusiaSavePictureResponse, err error) {
	uploadServer, err := vk.MarusiaGetPictureUploadLink(nil)
	if err != nil {
		return
	}

	bodyContent, err := vk.UploadFile(uploadServer.PictureUploadLink, file, "photo", "photo.jpg")
	if err != nil {
		return
	}

	var handler object.MarusiaPictureUploadResponse

	err = json.Unmarshal(bodyContent, &handler)
	if err != nil {
		return
	}

	photo, _ := json.Marshal(handler.Photo)

	response, err = vk.MarusiaSavePicture(Params{
		"server": handler.Server,
		"photo":  string(photo),
		"hash":   handler.Hash,
	})

	return
}

// UploadMarusiaAudio uploading audio.
//
// https://vk.com/dev/marusia_skill_docs10
func (vk *VK) UploadMarusiaAudio(file io.Reader) (response MarusiaCreateAudioResponse, err error) {
	uploadServer, err := vk.MarusiaGetAudioUploadLink(nil)
	if err != nil {
		return
	}

	bodyContent, err := vk.UploadFile(uploadServer.AudioUploadLink, file, "file", "audio.mp3")
	if err != nil {
		return
	}

	response, err = vk.MarusiaCreateAudio(Params{
		"audio_meta": string(bodyContent),
	})

	return
}
