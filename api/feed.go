package api

import (
	"context"
	"strings"

	"github.com/RasmusLindroth/go-mastodon"
	"github.com/RasmusLindroth/tut/config"
)

type TimelineType uint

func (ac *AccountClient) getStatusSimilar(fn func() ([]*mastodon.Status, error), timeline string) ([]Item, error) {
	var items []Item
	statuses, err := fn()
	if err != nil {
		return items, err
	}
	for _, s := range statuses {
		item := NewStatusItem(s, false)
		items = append(items, item)
	}
	return items, nil
}

func (ac *AccountClient) getUserSimilar(fn func() ([]*mastodon.Account, error), data interface{}) ([]Item, error) {
	var items []Item
	users, err := fn()
	if err != nil {
		return items, err
	}
	ids := []string{}
	for _, u := range users {
		ids = append(ids, string(u.ID))
	}
	rel, err := ac.Client.GetAccountRelationships(context.Background(), ids)
	if err != nil {
		return items, err
	}
	for _, u := range users {
		for _, r := range rel {
			if u.ID == r.ID {
				items = append(items, NewUserItem(&User{
					Data:           u,
					Relation:       r,
					AdditionalData: data,
				}, false))
				break
			}
		}
	}
	return items, nil
}

func (ac *AccountClient) GetTimeline(pg *mastodon.Pagination) ([]Item, error) {
	fn := func() ([]*mastodon.Status, error) {
		return ac.Client.GetTimelineHome(context.Background(), pg)
	}
	return ac.getStatusSimilar(fn, "home")
}

func (ac *AccountClient) GetTimelineFederated(pg *mastodon.Pagination) ([]Item, error) {
	fn := func() ([]*mastodon.Status, error) {
		return ac.Client.GetTimelinePublic(context.Background(), false, pg)
	}
	return ac.getStatusSimilar(fn, "public")
}

func (ac *AccountClient) GetTimelineLocal(pg *mastodon.Pagination) ([]Item, error) {
	fn := func() ([]*mastodon.Status, error) {
		return ac.Client.GetTimelinePublic(context.Background(), true, pg)
	}
	return ac.getStatusSimilar(fn, "public")
}

func (ac *AccountClient) GetNotifications(nth []config.NotificationToHide, pg *mastodon.Pagination) ([]Item, error) {
	var items []Item
	toHide := []string{}
	for _, n := range nth {
		toHide = append(toHide, string(n))
	}
	notifications, err := ac.Client.GetNotificationsExclude(context.Background(), &toHide, pg)
	if err != nil {
		return items, err
	}
	ids := []string{}
	for _, n := range notifications {
		ids = append(ids, string(n.Account.ID))
	}
	rel, err := ac.Client.GetAccountRelationships(context.Background(), ids)
	if err != nil {
		return items, err
	}
	for _, n := range notifications {
		for _, r := range rel {
			if n.Account.ID == r.ID {
				item := NewNotificationItem(n, &User{
					Data: &n.Account, Relation: r,
				})
				items = append(items, item)
				break
			}
		}
	}
	return items, nil
}

func (ac *AccountClient) GetHistory(status *mastodon.Status) ([]Item, error) {
	var items []Item
	statuses, err := ac.Client.GetStatusHistory(context.Background(), status.ID)
	if err != nil {
		return items, err
	}
	for _, s := range statuses {
		items = append(items, NewStatusHistoryItem(s))
	}
	return items, nil
}

func (ac *AccountClient) GetThread(status *mastodon.Status) ([]Item, error) {
	var items []Item
	statuses, err := ac.Client.GetStatusContext(context.Background(), status.ID)
	if err != nil {
		return items, err
	}
	for _, s := range statuses.Ancestors {
		items = append(items, NewStatusItem(s, false))
	}
	items = append(items, NewStatusItem(status, false))
	for _, s := range statuses.Descendants {
		items = append(items, NewStatusItem(s, false))
	}
	return items, nil
}

func (ac *AccountClient) GetFavorites(pg *mastodon.Pagination) ([]Item, error) {
	fn := func() ([]*mastodon.Status, error) {
		return ac.Client.GetFavourites(context.Background(), pg)
	}
	return ac.getStatusSimilar(fn, "home")
}

func (ac *AccountClient) GetBookmarks(pg *mastodon.Pagination) ([]Item, error) {
	fn := func() ([]*mastodon.Status, error) {
		return ac.Client.GetBookmarks(context.Background(), pg)
	}
	return ac.getStatusSimilar(fn, "home")
}

func (ac *AccountClient) GetConversations(pg *mastodon.Pagination) ([]Item, error) {
	var items []Item
	conversations, err := ac.Client.GetConversations(context.Background(), pg)
	if err != nil {
		return items, err
	}
	for _, c := range conversations {
		item := NewStatusItem(c.LastStatus, false)
		items = append(items, item)
	}
	return items, nil
}

func (ac *AccountClient) GetUsers(search string) ([]Item, error) {
	var items []Item
	var users []*mastodon.Account
	var err error
	if strings.HasPrefix(search, "@") && len(strings.Split(search, "@")) == 3 {
		users, err = ac.Client.AccountsSearch(context.Background(), search, 10, true)
	}
	if len(users) == 0 || err != nil {
		users, err = ac.Client.AccountsSearch(context.Background(), search, 10, false)
	}
	if err != nil {
		return items, err
	}
	ids := []string{}
	for _, u := range users {
		ids = append(ids, string(u.ID))
	}
	rel, err := ac.Client.GetAccountRelationships(context.Background(), ids)
	if err != nil {
		return items, err
	}
	for _, u := range users {
		for _, r := range rel {
			if u.ID == r.ID {
				items = append(items, NewUserItem(&User{
					Data:     u,
					Relation: r,
				}, false))
				break
			}
		}
	}
	return items, nil
}

func (ac *AccountClient) GetBoostsStatus(pg *mastodon.Pagination, id mastodon.ID) ([]Item, error) {
	fn := func() ([]*mastodon.Account, error) {
		return ac.Client.GetRebloggedBy(context.Background(), id, pg)
	}
	return ac.getUserSimilar(fn, nil)
}

func (ac *AccountClient) GetFavoritesStatus(pg *mastodon.Pagination, id mastodon.ID) ([]Item, error) {
	fn := func() ([]*mastodon.Account, error) {
		return ac.Client.GetFavouritedBy(context.Background(), id, pg)
	}
	return ac.getUserSimilar(fn, nil)
}

func (ac *AccountClient) GetFollowers(pg *mastodon.Pagination, id mastodon.ID) ([]Item, error) {
	fn := func() ([]*mastodon.Account, error) {
		return ac.Client.GetAccountFollowers(context.Background(), id, pg)
	}
	return ac.getUserSimilar(fn, nil)
}

func (ac *AccountClient) GetFollowing(pg *mastodon.Pagination, id mastodon.ID) ([]Item, error) {
	fn := func() ([]*mastodon.Account, error) {
		return ac.Client.GetAccountFollowing(context.Background(), id, pg)
	}
	return ac.getUserSimilar(fn, nil)
}

func (ac *AccountClient) GetBlocking(pg *mastodon.Pagination) ([]Item, error) {
	fn := func() ([]*mastodon.Account, error) {
		return ac.Client.GetBlocks(context.Background(), pg)
	}
	return ac.getUserSimilar(fn, nil)
}

func (ac *AccountClient) GetMuting(pg *mastodon.Pagination) ([]Item, error) {
	fn := func() ([]*mastodon.Account, error) {
		return ac.Client.GetMutes(context.Background(), pg)
	}
	return ac.getUserSimilar(fn, nil)
}

func (ac *AccountClient) GetFollowRequests(pg *mastodon.Pagination) ([]Item, error) {
	fn := func() ([]*mastodon.Account, error) {
		return ac.Client.GetFollowRequests(context.Background(), pg)
	}
	return ac.getUserSimilar(fn, nil)
}

func (ac *AccountClient) GetUser(pg *mastodon.Pagination, id mastodon.ID) ([]Item, error) {
	var items []Item
	statuses, err := ac.Client.GetAccountStatuses(context.Background(), id, pg)
	if err != nil {
		return items, err
	}
	for _, s := range statuses {
		item := NewStatusItem(s, false)
		items = append(items, item)
	}
	return items, nil
}

func (ac *AccountClient) GetUserPinned(id mastodon.ID) ([]Item, error) {
	var items []Item
	statuses, err := ac.Client.GetAccountPinnedStatuses(context.Background(), id)
	if err != nil {
		return items, err
	}
	for _, s := range statuses {
		item := NewStatusItem(s, true)
		items = append(items, item)
	}
	return items, nil
}

func (ac *AccountClient) GetTags(pg *mastodon.Pagination) ([]Item, error) {
	var items []Item
	tags, err := ac.Client.TagsFollowed(context.Background(), pg)
	if err != nil {
		return items, err
	}
	for _, t := range tags {
		items = append(items, NewTagItem(t))
	}
	return items, nil
}

func (ac *AccountClient) GetLists() ([]Item, error) {
	var items []Item
	lists, err := ac.Client.GetLists(context.Background())
	if err != nil {
		return items, err
	}
	for _, l := range lists {
		items = append(items, NewListsItem(l))
	}
	return items, nil
}

func (ac *AccountClient) GetListStatuses(pg *mastodon.Pagination, id mastodon.ID) ([]Item, error) {
	var items []Item
	statuses, err := ac.Client.GetTimelineList(context.Background(), id, pg)
	if err != nil {
		return items, err
	}
	for _, s := range statuses {
		item := NewStatusItem(s, false)
		items = append(items, item)
	}
	return items, nil
}

func (ac *AccountClient) GetFollowingForList(pg *mastodon.Pagination, id mastodon.ID, data interface{}) ([]Item, error) {
	fn := func() ([]*mastodon.Account, error) {
		return ac.Client.GetAccountFollowing(context.Background(), id, pg)
	}
	return ac.getUserSimilar(fn, data)
}

func (ac *AccountClient) GetListUsers(pg *mastodon.Pagination, id mastodon.ID, data interface{}) ([]Item, error) {
	fn := func() ([]*mastodon.Account, error) {
		return ac.Client.GetListAccounts(context.Background(), id)
	}
	return ac.getUserSimilar(fn, data)
}

func (ac *AccountClient) GetTag(pg *mastodon.Pagination, search string) ([]Item, error) {
	fn := func() ([]*mastodon.Status, error) {
		return ac.Client.GetTimelineHashtag(context.Background(), search, false, pg)
	}
	return ac.getStatusSimilar(fn, "public")
}

func (ac *AccountClient) GetTagMultiple(pg *mastodon.Pagination, search string) ([]Item, error) {
	fn := func() ([]*mastodon.Status, error) {
		var s string
		td := mastodon.TagData{}
		parts := strings.Split(search, " ")
		for i, p := range parts {
			if i == 0 {
				s = p
				continue
			}
			if len(p) > 0 {
				td.Any = append(td.Any, p)
			}
		}
		return ac.Client.GetTimelineHashtagMultiple(context.Background(), s, false, &td, pg)
	}
	return ac.getStatusSimilar(fn, "public")
}
