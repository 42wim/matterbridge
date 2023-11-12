package communities

import (
	"sort"

	"github.com/status-im/status-go/protocol/protobuf"
)

func (o *Community) ChatsByCategoryID(categoryID string) []string {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	var chatIDs []string
	if o.config == nil || o.config.CommunityDescription == nil {
		return chatIDs
	}

	for chatID, chat := range o.config.CommunityDescription.Chats {
		if chat.CategoryId == categoryID {
			chatIDs = append(chatIDs, chatID)
		}
	}
	return chatIDs
}

func (o *Community) CommunityChatsIDs() []string {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	var chatIDs []string
	if o.config == nil || o.config.CommunityDescription == nil {
		return chatIDs
	}

	for chatID := range o.config.CommunityDescription.Chats {
		chatIDs = append(chatIDs, chatID)
	}
	return chatIDs
}

func (o *Community) CreateCategory(categoryID string, categoryName string, chatIDs []string) (*CommunityChanges, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !(o.IsControlNode() || o.hasPermissionToSendCommunityEvent(protobuf.CommunityEvent_COMMUNITY_CATEGORY_CREATE)) {
		return nil, ErrNotAuthorized
	}

	changes, err := o.createCategory(categoryID, categoryName, chatIDs)
	if err != nil {
		return nil, err
	}

	changes.CategoriesAdded[categoryID] = o.config.CommunityDescription.Categories[categoryID]
	for i, cid := range chatIDs {
		changes.ChatsModified[cid] = &CommunityChatChanges{
			MembersAdded:     make(map[string]*protobuf.CommunityMember),
			MembersRemoved:   make(map[string]*protobuf.CommunityMember),
			CategoryModified: categoryID,
			PositionModified: i,
		}
	}

	if o.IsControlNode() {
		o.increaseClock()
	} else {
		err := o.addNewCommunityEvent(o.ToCreateCategoryCommunityEvent(categoryID, categoryName, chatIDs))
		if err != nil {
			return nil, err
		}
	}

	return changes, nil
}

func (o *Community) EditCategory(categoryID string, categoryName string, chatIDs []string) (*CommunityChanges, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !(o.IsControlNode() || o.hasPermissionToSendCommunityEvent(protobuf.CommunityEvent_COMMUNITY_CATEGORY_EDIT)) {
		return nil, ErrNotAuthorized
	}

	changes, err := o.editCategory(categoryID, categoryName, chatIDs)
	if err != nil {
		return nil, err
	}

	changes.CategoriesModified[categoryID] = o.config.CommunityDescription.Categories[categoryID]
	for i, cid := range chatIDs {
		changes.ChatsModified[cid] = &CommunityChatChanges{
			MembersAdded:     make(map[string]*protobuf.CommunityMember),
			MembersRemoved:   make(map[string]*protobuf.CommunityMember),
			CategoryModified: categoryID,
			PositionModified: i,
		}
	}

	if o.IsControlNode() {
		o.increaseClock()
	} else {
		err := o.addNewCommunityEvent(o.ToEditCategoryCommunityEvent(categoryID, categoryName, chatIDs))
		if err != nil {
			return nil, err
		}
	}

	return changes, nil
}

func (o *Community) ReorderCategories(categoryID string, newPosition int) (*CommunityChanges, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !(o.IsControlNode() || o.hasPermissionToSendCommunityEvent(protobuf.CommunityEvent_COMMUNITY_CATEGORY_REORDER)) {
		return nil, ErrNotAuthorized
	}

	changes, err := o.reorderCategories(categoryID, newPosition)
	if err != nil {
		return nil, err
	}

	if o.IsControlNode() {
		o.increaseClock()
	} else {
		err := o.addNewCommunityEvent(o.ToReorderCategoryCommunityEvent(categoryID, newPosition))
		if err != nil {
			return nil, err
		}
	}

	return changes, nil
}

func (o *Community) setModifiedCategories(changes *CommunityChanges, s sortSlice) {
	sort.Sort(s)
	for i, catSortHelper := range s {
		if o.config.CommunityDescription.Categories[catSortHelper.catID].Position != int32(i) {
			o.config.CommunityDescription.Categories[catSortHelper.catID].Position = int32(i)
			changes.CategoriesModified[catSortHelper.catID] = o.config.CommunityDescription.Categories[catSortHelper.catID]
		}
	}
}

func (o *Community) ReorderChat(categoryID string, chatID string, newPosition int) (*CommunityChanges, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !(o.IsControlNode() || o.hasPermissionToSendCommunityEvent(protobuf.CommunityEvent_COMMUNITY_CHANNEL_REORDER)) {
		return nil, ErrNotAuthorized
	}

	changes, err := o.reorderChat(categoryID, chatID, newPosition)
	if err != nil {
		return nil, err
	}

	if o.IsControlNode() {
		o.increaseClock()
	} else {
		err := o.addNewCommunityEvent(o.ToReorderChannelCommunityEvent(categoryID, chatID, newPosition))
		if err != nil {
			return nil, err
		}
	}

	return changes, nil
}

func (o *Community) SortCategoryChats(changes *CommunityChanges, categoryID string) {
	var catChats []string
	for k, c := range o.config.CommunityDescription.Chats {
		if c.CategoryId == categoryID {
			catChats = append(catChats, k)
		}
	}

	sortedChats := make(sortSlice, 0, len(catChats))
	for _, k := range catChats {
		sortedChats = append(sortedChats, sorterHelperIdx{
			pos:    o.config.CommunityDescription.Chats[k].Position,
			chatID: k,
		})
	}

	sort.Sort(sortedChats)

	for i, chatSortHelper := range sortedChats {
		if o.config.CommunityDescription.Chats[chatSortHelper.chatID].Position != int32(i) {
			o.config.CommunityDescription.Chats[chatSortHelper.chatID].Position = int32(i)
			if changes.ChatsModified[chatSortHelper.chatID] != nil {
				changes.ChatsModified[chatSortHelper.chatID].PositionModified = i
			} else {
				changes.ChatsModified[chatSortHelper.chatID] = &CommunityChatChanges{
					PositionModified: i,
					MembersAdded:     make(map[string]*protobuf.CommunityMember),
					MembersRemoved:   make(map[string]*protobuf.CommunityMember),
				}
			}
		}
	}
}

func (o *Community) insertAndSort(changes *CommunityChanges, oldCategoryID string, categoryID string, chatID string, chat *protobuf.CommunityChat, newPosition int) {
	// We sort the chats here because maps are not guaranteed to keep order
	var catChats []string
	sortedChats := make(sortSlice, 0, len(o.config.CommunityDescription.Chats))
	for k, v := range o.config.CommunityDescription.Chats {
		sortedChats = append(sortedChats, sorterHelperIdx{
			pos:    v.Position,
			chatID: k,
		})
	}
	sort.Sort(sortedChats)
	for _, k := range sortedChats {
		if o.config.CommunityDescription.Chats[k.chatID].CategoryId == categoryID {
			catChats = append(catChats, k.chatID)
		}
	}

	if newPosition > 0 && newPosition >= len(catChats) {
		newPosition = len(catChats) - 1
	} else if newPosition < 0 {
		newPosition = 0
	}

	decrease := false
	if chat.Position > int32(newPosition) {
		decrease = true
	}

	for k, v := range o.config.CommunityDescription.Chats {
		if k != chatID && newPosition == int(v.Position) && v.CategoryId == categoryID {
			if oldCategoryID == categoryID {
				if decrease {
					v.Position++
				} else {
					v.Position--
				}
			} else {
				v.Position++
			}
		}
	}

	idx := -1
	currChatID := ""
	var sortedChatIDs []string
	for i, k := range catChats {
		if o.config.CommunityDescription.Chats[k] != chat && ((decrease && o.config.CommunityDescription.Chats[k].Position < int32(newPosition)) || (!decrease && o.config.CommunityDescription.Chats[k].Position <= int32(newPosition))) {
			sortedChatIDs = append(sortedChatIDs, k)
		} else {
			if o.config.CommunityDescription.Chats[k] == chat {
				idx = i
				currChatID = k
			}
		}
	}

	sortedChatIDs = append(sortedChatIDs, currChatID)

	for i, k := range catChats {
		if i == idx || (decrease && o.config.CommunityDescription.Chats[k].Position < int32(newPosition)) || (!decrease && o.config.CommunityDescription.Chats[k].Position <= int32(newPosition)) {
			continue
		}
		sortedChatIDs = append(sortedChatIDs, k)
	}

	for i, sortedChatID := range sortedChatIDs {
		if o.config.CommunityDescription.Chats[sortedChatID].Position != int32(i) {
			o.config.CommunityDescription.Chats[sortedChatID].Position = int32(i)
			if changes.ChatsModified[sortedChatID] != nil {
				changes.ChatsModified[sortedChatID].PositionModified = i
			} else {
				changes.ChatsModified[sortedChatID] = &CommunityChatChanges{
					MembersAdded:     make(map[string]*protobuf.CommunityMember),
					MembersRemoved:   make(map[string]*protobuf.CommunityMember),
					PositionModified: i,
				}
			}
		}
	}
}

func (o *Community) getCategoryChatCount(categoryID string) int {
	result := 0
	for _, chat := range o.config.CommunityDescription.Chats {
		if chat.CategoryId == categoryID {
			result = result + 1
		}
	}
	return result
}

func (o *Community) DeleteCategory(categoryID string) (*CommunityChanges, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !(o.IsControlNode() || o.hasPermissionToSendCommunityEvent(protobuf.CommunityEvent_COMMUNITY_CATEGORY_DELETE)) {
		return nil, ErrNotAuthorized
	}

	changes, err := o.deleteCategory(categoryID)
	if err != nil {
		return nil, err
	}

	if o.IsControlNode() {
		o.increaseClock()
	} else {
		err := o.addNewCommunityEvent(o.ToDeleteCategoryCommunityEvent(categoryID))
		if err != nil {
			return nil, err
		}
	}

	return changes, nil
}

func (o *Community) createCategory(categoryID string, categoryName string, chatIDs []string) (*CommunityChanges, error) {
	if o.config.CommunityDescription.Categories == nil {
		o.config.CommunityDescription.Categories = make(map[string]*protobuf.CommunityCategory)
	}
	if _, ok := o.config.CommunityDescription.Categories[categoryID]; ok {
		return nil, ErrCategoryAlreadyExists
	}

	for _, cid := range chatIDs {
		c, exists := o.config.CommunityDescription.Chats[cid]
		if !exists {
			return nil, ErrChatNotFound
		}

		if exists && c.CategoryId != categoryID && c.CategoryId != "" {
			return nil, ErrChatAlreadyAssigned
		}
	}

	changes := o.emptyCommunityChanges()

	o.config.CommunityDescription.Categories[categoryID] = &protobuf.CommunityCategory{
		CategoryId: categoryID,
		Name:       categoryName,
		Position:   int32(len(o.config.CommunityDescription.Categories)),
	}

	for i, cid := range chatIDs {
		o.config.CommunityDescription.Chats[cid].CategoryId = categoryID
		o.config.CommunityDescription.Chats[cid].Position = int32(i)
	}

	o.SortCategoryChats(changes, "")

	return changes, nil
}

func (o *Community) editCategory(categoryID string, categoryName string, chatIDs []string) (*CommunityChanges, error) {
	if o.config.CommunityDescription.Categories == nil {
		o.config.CommunityDescription.Categories = make(map[string]*protobuf.CommunityCategory)
	}
	if _, ok := o.config.CommunityDescription.Categories[categoryID]; !ok {
		return nil, ErrCategoryNotFound
	}

	for _, cid := range chatIDs {
		c, exists := o.config.CommunityDescription.Chats[cid]
		if !exists {
			return nil, ErrChatNotFound
		}

		if exists && c.CategoryId != categoryID && c.CategoryId != "" {
			return nil, ErrChatAlreadyAssigned
		}
	}

	changes := o.emptyCommunityChanges()

	emptyCatLen := o.getCategoryChatCount("")

	// remove any chat that might have been assigned before and now it's not part of the category
	var chatsToRemove []string
	for k, chat := range o.config.CommunityDescription.Chats {
		if chat.CategoryId == categoryID {
			found := false
			for _, c := range chatIDs {
				if k == c {
					found = true
				}
			}
			if !found {
				chat.CategoryId = ""
				chatsToRemove = append(chatsToRemove, k)
			}
		}
	}

	o.config.CommunityDescription.Categories[categoryID].Name = categoryName

	for i, cid := range chatIDs {
		o.config.CommunityDescription.Chats[cid].CategoryId = categoryID
		o.config.CommunityDescription.Chats[cid].Position = int32(i)
	}

	for i, cid := range chatsToRemove {
		o.config.CommunityDescription.Chats[cid].Position = int32(emptyCatLen + i)
		changes.ChatsModified[cid] = &CommunityChatChanges{
			MembersAdded:     make(map[string]*protobuf.CommunityMember),
			MembersRemoved:   make(map[string]*protobuf.CommunityMember),
			CategoryModified: "",
			PositionModified: int(o.config.CommunityDescription.Chats[cid].Position),
		}
	}

	o.SortCategoryChats(changes, "")

	return changes, nil
}

func (o *Community) deleteCategory(categoryID string) (*CommunityChanges, error) {
	if _, exists := o.config.CommunityDescription.Categories[categoryID]; !exists {
		return nil, ErrCategoryNotFound
	}

	changes := o.emptyCommunityChanges()

	emptyCategoryChatCount := o.getCategoryChatCount("")
	i := 0
	for _, chat := range o.config.CommunityDescription.Chats {
		if chat.CategoryId == categoryID {
			i++
			chat.CategoryId = ""
			chat.Position = int32(emptyCategoryChatCount + i)
		}
	}

	o.SortCategoryChats(changes, "")

	delete(o.config.CommunityDescription.Categories, categoryID)

	changes.CategoriesRemoved = append(changes.CategoriesRemoved, categoryID)

	// Reorder
	s := make(sortSlice, 0, len(o.config.CommunityDescription.Categories))
	for _, cat := range o.config.CommunityDescription.Categories {
		s = append(s, sorterHelperIdx{
			pos:   cat.Position,
			catID: cat.CategoryId,
		})
	}

	o.setModifiedCategories(changes, s)

	return changes, nil
}

func (o *Community) reorderCategories(categoryID string, newPosition int) (*CommunityChanges, error) {
	if _, exists := o.config.CommunityDescription.Categories[categoryID]; !exists {
		return nil, ErrCategoryNotFound
	}

	if newPosition > 0 && newPosition >= len(o.config.CommunityDescription.Categories) {
		newPosition = len(o.config.CommunityDescription.Categories) - 1
	} else if newPosition < 0 {
		newPosition = 0
	}

	category := o.config.CommunityDescription.Categories[categoryID]
	if category.Position == int32(newPosition) {
		return nil, ErrNoChangeInPosition
	}

	decrease := false
	if category.Position > int32(newPosition) {
		decrease = true
	}

	// Sorting the categories because maps are not guaranteed to keep order
	s := make(sortSlice, 0, len(o.config.CommunityDescription.Categories))
	for k, v := range o.config.CommunityDescription.Categories {
		s = append(s, sorterHelperIdx{
			pos:   v.Position,
			catID: k,
		})
	}
	sort.Sort(s)
	var communityCategories []*protobuf.CommunityCategory
	for _, currCat := range s {
		communityCategories = append(communityCategories, o.config.CommunityDescription.Categories[currCat.catID])
	}

	var sortedCategoryIDs []string
	for _, v := range communityCategories {
		if v != category && ((decrease && v.Position < int32(newPosition)) || (!decrease && v.Position <= int32(newPosition))) {
			sortedCategoryIDs = append(sortedCategoryIDs, v.CategoryId)
		}
	}

	sortedCategoryIDs = append(sortedCategoryIDs, categoryID)

	for _, v := range communityCategories {
		if v.CategoryId == categoryID || (decrease && v.Position < int32(newPosition)) || (!decrease && v.Position <= int32(newPosition)) {
			continue
		}
		sortedCategoryIDs = append(sortedCategoryIDs, v.CategoryId)
	}

	s = make(sortSlice, 0, len(o.config.CommunityDescription.Categories))
	for i, k := range sortedCategoryIDs {
		s = append(s, sorterHelperIdx{
			pos:   int32(i),
			catID: k,
		})
	}

	changes := o.emptyCommunityChanges()

	o.setModifiedCategories(changes, s)

	return changes, nil
}

func (o *Community) reorderChat(categoryID string, chatID string, newPosition int) (*CommunityChanges, error) {
	if categoryID != "" {
		if _, exists := o.config.CommunityDescription.Categories[categoryID]; !exists {
			return nil, ErrCategoryNotFound
		}
	}

	var chat *protobuf.CommunityChat
	var exists bool
	if chat, exists = o.config.CommunityDescription.Chats[chatID]; !exists {
		return nil, ErrChatNotFound
	}

	oldCategoryID := chat.CategoryId
	chat.CategoryId = categoryID

	changes := o.emptyCommunityChanges()

	o.SortCategoryChats(changes, oldCategoryID)
	o.insertAndSort(changes, oldCategoryID, categoryID, chatID, chat, newPosition)

	return changes, nil
}
