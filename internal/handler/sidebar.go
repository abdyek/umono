package handler

import (
	"github.com/umono-cms/umono/internal/service"
	"github.com/umono-cms/umono/internal/view"
)

// sitePageList returns the page of the site page sidebar at offset.
func sitePageList(s *service.SitePageService, offset int, activeID uint) view.SitePageList {
	items, next := s.GetPage(offset)
	return view.SitePageList{Items: view.SitePageUl(items, activeID), Next: next, ActiveID: activeID}
}

// componentList returns the page of the component sidebar at offset.
func componentList(s *service.ComponentService, offset int, activeID uint) view.ComponentList {
	items, next := s.GetPage(offset)
	return view.ComponentList{Items: view.ComponentUl(items, activeID), Next: next, ActiveID: activeID}
}

// mediaSidebar returns the page of the media sidebar at offset.
func (h *mediaHandler) mediaSidebar(offset int, activeID string) view.MediaSidebar {
	items, next := h.mediaService.GetPage(offset)
	return view.MediaSidebar{Items: h.buildMediaList(items, activeID), Next: next, ActiveID: activeID}
}
