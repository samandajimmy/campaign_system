package rewards

import (
	"gade/srv-gade-point/models"

	"github.com/labstack/echo"
)

// Repository represent the rewards repository contract
type Repository interface {
	CreateReward(echo.Context, *models.Reward, int64) error
	DeleteByCampaign(echo.Context, int64) error
	CreateRewardTag(echo.Context, *models.Tag, int64) error
	DeleteRewardTag(echo.Context, int64) error
	GetRewardByCampaign(echo.Context, int64) ([]models.Reward, error)
	GetRewardTags(echo.Context, *models.Reward) (*models.Reward, error)
}
