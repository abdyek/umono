package reqbodies

type MDToHTML struct {
	MD string `json:"markdown"`
}

type UmonoLangToHTML struct {
	UmonoLang string `json:"umono_lang"`
}
