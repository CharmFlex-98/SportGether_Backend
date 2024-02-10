package remote_config

import (
	"sportgether/constants"
	"sportgether/utils"
)

type SportDetails struct {
	Sports []Sport `json:"sports"`
}

type Sport struct {
	Index    int64  `json:"sportIndex"`
	Sport    string `json:"sport"`
	ImageUrl string `json:"imageUrl"`
}

func FromSportToImageUrl(sport string) (string, error) {
	sportDetail := SportDetails{}
	err := utils.ReadJsonFromFile("./data/available_sports_detail.json", &sportDetail)
	if err != nil {
		return "", err
	}

	for _, value := range sportDetail.Sports {
		if value.Sport == sport {
			return value.ImageUrl, nil
		}
	}

	return "", constants.SportConfigNotFoundError
}
