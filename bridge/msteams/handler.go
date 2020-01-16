package bmsteams

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"

	msgraph "github.com/yaegashi/msgraph.go/beta"
)

func (b *Bmsteams) findFile(weburl string) (string, error) {
	itemRB, err := b.gc.GetDriveItemByURL(b.ctx, weburl)
	if err != nil {
		return "", err
	}
	itemRB.Workbook().Worksheets()
	b.gc.Workbooks()
	item, err := itemRB.Request().Get(b.ctx)
	if err != nil {
		return "", err
	}
	if url, ok := item.GetAdditionalData("@microsoft.graph.downloadUrl"); ok {
		return url.(string), nil
	}
	return "", nil
}

// handleDownloadFile handles file download
func (b *Bmsteams) handleDownloadFile(rmsg *config.Message, filename, weburl string) error {
	realURL, err := b.findFile(weburl)
	if err != nil {
		return err
	}
	// Actually download the file.
	data, err := helper.DownloadFile(realURL)
	if err != nil {
		return fmt.Errorf("download %s failed %#v", weburl, err)
	}

	// If a comment is attached to the file(s) it is in the 'Text' field of the teams messge event
	// and should be added as comment to only one of the files. We reset the 'Text' field to ensure
	// that the comment is not duplicated.
	comment := rmsg.Text
	rmsg.Text = ""
	helper.HandleDownloadData(b.Log, rmsg, filename, comment, weburl, data, b.General)
	return nil
}

func (b *Bmsteams) handleAttachments(rmsg *config.Message, msg msgraph.ChatMessage) {
	for _, a := range msg.Attachments {
		//remove the attachment tags from the text
		rmsg.Text = attachRE.ReplaceAllString(rmsg.Text, "")

		//handle a code snippet (code block)
		if *a.ContentType == "application/vnd.microsoft.card.codesnippet" {
			b.handleCodeSnippet(rmsg, a)
			continue
		}

		//handle the download
		err := b.handleDownloadFile(rmsg, *a.Name, *a.ContentURL)
		if err != nil {
			b.Log.Errorf("download of %s failed: %s", *a.Name, err)
		}
	}
}

type AttachContent struct {
	Language       string `json:"language"`
	CodeSnippetURL string `json:"codeSnippetUrl"`
}

func (b *Bmsteams) handleCodeSnippet(rmsg *config.Message, attach msgraph.ChatMessageAttachment) {
	var content AttachContent
	err := json.Unmarshal([]byte(*attach.Content), &content)
	if err != nil {
		b.Log.Errorf("unmarshal codesnippet failed: %s", err)
		return
	}
	s := strings.Split(content.CodeSnippetURL, "/")
	if len(s) != 13 {
		b.Log.Errorf("codesnippetUrl has unexpected size: %s", content.CodeSnippetURL)
		return
	}
	resp, err := b.gc.Teams().Request().Client().Get(content.CodeSnippetURL)
	if err != nil {
		b.Log.Errorf("retrieving snippet content failed:%s", err)
		return
	}
	defer resp.Body.Close()
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		b.Log.Errorf("reading snippet data failed: %s", err)
		return
	}
	rmsg.Text = rmsg.Text + "\n```" + content.Language + "\n" + string(res) + "\n```\n"
}
