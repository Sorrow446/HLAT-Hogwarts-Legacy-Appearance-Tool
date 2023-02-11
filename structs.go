package main

type Args struct {
	Command   string `arg:"positional, required" help:"import or export"`
	InPath    string `arg:"-i, required" help:"Path of input file. JSON appearance if import, save file if export."`
	OutPath   string `arg:"-o" help:"Path of output file. Save file if import, JSON appearance if export."`
	OrigName  bool   `arg:"--orig-name" help:"Keep original character name."`
}

type AppearanceData struct {
	PresetType string `json:"presetType"`
	PresetName string `json:"presetName"`
}

type GenderData struct {
	DataName string `json:"dataName"`
	DataValue string `json:"dataValue"`
}

type Appearance struct {
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	AppearanceData []*AppearanceData `json:"appearanceData"`
	GenderData     []*GenderData `json:"genderData"`
}