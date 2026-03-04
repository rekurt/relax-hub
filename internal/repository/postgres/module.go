package postgres

import (
	"github.com/nikitaaldaev/bani/internal/repository"
	"go.uber.org/fx"
)

var Module = fx.Module("repository",
	fx.Provide(
		fx.Annotate(NewUserRepository, fx.As(new(repository.UserRepository))),
		fx.Annotate(NewCityRepository, fx.As(new(repository.CityRepository))),
		fx.Annotate(NewBathhouseRepository, fx.As(new(repository.BathhouseRepository))),
		fx.Annotate(NewBookingRepository, fx.As(new(repository.BookingRepository))),
		fx.Annotate(NewReviewRepository, fx.As(new(repository.ReviewRepository))),
		fx.Annotate(NewFavoriteRepository, fx.As(new(repository.FavoriteRepository))),
		fx.Annotate(NewRepresentativeRepository, fx.As(new(repository.RepresentativeRepository))),
		fx.Annotate(NewNotificationRepository, fx.As(new(repository.NotificationRepository))),
	),
)
