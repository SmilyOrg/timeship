package api

import "net/http"

func (s *Server) PostArchive(w http.ResponseWriter, r *http.Request, params PostArchiveParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) PostDelete(w http.ResponseWriter, r *http.Request, params PostDeleteParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) GetDownload(w http.ResponseWriter, r *http.Request, params GetDownloadParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) PostMove(w http.ResponseWriter, r *http.Request, params PostMoveParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) PostNewfile(w http.ResponseWriter, r *http.Request, params PostNewfileParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) PostNewfolder(w http.ResponseWriter, r *http.Request, params PostNewfolderParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) GetPreview(w http.ResponseWriter, r *http.Request, params GetPreviewParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) PostRename(w http.ResponseWriter, r *http.Request, params PostRenameParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) PostSave(w http.ResponseWriter, r *http.Request, params PostSaveParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) GetSearch(w http.ResponseWriter, r *http.Request, params GetSearchParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) GetSubfolders(w http.ResponseWriter, r *http.Request, params GetSubfoldersParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) PostUnarchive(w http.ResponseWriter, r *http.Request, params PostUnarchiveParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) PostUpload(w http.ResponseWriter, r *http.Request, params PostUploadParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}
