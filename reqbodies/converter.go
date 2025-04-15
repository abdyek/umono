package reqbodies

type MDToHTML struct {
	MD string `json:"markdown"`
}

type UmonoLangToHTML struct {
	UmonoLang string `json:"umono_lang"`
}

type UmonoLangToHTMLForGlobalComp struct {
	CompName string `json:"component_name"`
	Content  string `json:"content"`
}
