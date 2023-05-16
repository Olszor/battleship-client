package http

type InitGameRequest struct {
	Desc       string `json:"desc"`
	Nick       string `json:"nick"`
	TargetNick string `json:"target_nick"`
	Wpbot      bool   `json:"wpbot"`
}

type BoardResponse struct {
	Board []string `json:"board"`
}

type StatusResponse struct {
	GameStatus     string   `json:"game_status"`
	LastGameStatus string   `json:"last_game_status"`
	Nick           string   `json:"nick"`
	OppShots       []string `json:"opp_shots"`
	Opponent       string   `json:"opponent"`
	ShouldFire     bool     `json:"should_fire"`
	Timer          int      `json:"timer"`
}

type DescriptionResponse struct {
	Desc     string `json:"desc"`
	Nick     string `json:"nick"`
	OppDesc  string `json:"opp_desc"`
	Opponent string `json:"opponent"`
}

type FireRequest struct {
	Coord string `json:"coord"`
}

type FireResponse struct {
	Result string `json:"result"`
}

type ListResponse struct {
	GameStatus string `json:"game_status"`
	Nick       string `json:"nick"`
}
