package api

import (
	"net/http"
	"path/filepath"

	"github.com/elgatito/elementum/api/repository"
	"github.com/elgatito/elementum/bittorrent"
	"github.com/elgatito/elementum/config"
	"github.com/elgatito/elementum/providers"

	"github.com/gin-gonic/gin"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("api")

// Routes ...
func Routes(btService *bittorrent.BTService) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithWriter(gin.DefaultWriter, "/torrents/list", "/notification"))

	gin.SetMode(gin.ReleaseMode)

	r.GET("/", Index)
	r.GET("/search", Search(btService))
	r.GET("/playtorrent", PlayTorrent)
	r.GET("/infolabels", InfoLabelsStored(btService))
	r.GET("/changelog", Changelog)

	r.LoadHTMLGlob(filepath.Join(config.Get().Info.Path, "resources", "web", "*.html"))
	web := r.Group("/web")
	{
		web.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", nil)
		})
		web.Static("/static", filepath.Join(config.Get().Info.Path, "resources", "web", "static"))
		web.StaticFile("/favicon.ico", filepath.Join(config.Get().Info.Path, "resources", "web", "favicon.ico"))
	}

	torrents := r.Group("/torrents")
	{
		torrents.GET("/", ListTorrents(btService))
		torrents.GET("/add", AddTorrent(btService))
		torrents.GET("/pause", PauseSession(btService))
		torrents.GET("/resume", ResumeSession(btService))
		torrents.GET("/move/:torrentId", MoveTorrent(btService))
		torrents.GET("/pause/:torrentId", PauseTorrent(btService))
		torrents.GET("/resume/:torrentId", ResumeTorrent(btService))
		torrents.GET("/delete/:torrentId", RemoveTorrent(btService))

		// Web UI json
		torrents.GET("/list", ListTorrentsWeb(btService))
	}

	movies := r.Group("/movies")
	{
		movies.GET("/", MoviesIndex)
		movies.GET("/search", SearchMovies)
		movies.GET("/popular", PopularMovies)
		movies.GET("/popular/:genre", PopularMovies)
		movies.GET("/recent", RecentMovies)
		movies.GET("/recent/:genre", RecentMovies)
		movies.GET("/top", TopRatedMovies)
		movies.GET("/imdb250", IMDBTop250)
		movies.GET("/mostvoted", MoviesMostVoted)
		movies.GET("/genres", MovieGenres)

		trakt := movies.Group("/trakt")
		{
			trakt.GET("/watchlist", WatchlistMovies)
			trakt.GET("/collection", CollectionMovies)
			trakt.GET("/popular", TraktPopularMovies)
			trakt.GET("/trending", TraktTrendingMovies)
			trakt.GET("/played", TraktMostPlayedMovies)
			trakt.GET("/watched", TraktMostWatchedMovies)
			trakt.GET("/collected", TraktMostCollectedMovies)
			trakt.GET("/anticipated", TraktMostAnticipatedMovies)
			trakt.GET("/boxoffice", TraktBoxOffice)
			trakt.GET("/history", TraktHistoryMovies)

			lists := trakt.Group("/lists")
			{
				lists.GET("/", MoviesTraktLists)
				lists.GET("/id/:listId", UserlistMovies)
			}

			calendars := trakt.Group("/calendars")
			{
				calendars.GET("/", CalendarMovies)
				calendars.GET("/movies", TraktMyMovies)
				calendars.GET("/releases", TraktMyReleases)
				calendars.GET("/allmovies", TraktAllMovies)
				calendars.GET("/allreleases", TraktAllReleases)
			}
		}
	}
	movie := r.Group("/movie")
	{
		movie.GET("/:tmdbId/infolabels", InfoLabelsMovie(btService))
		movie.GET("/:tmdbId/links", MoviePlaySelector("links", btService))
		movie.GET("/:tmdbId/forcelinks", MoviePlaySelector("forcelinks", btService))
		movie.GET("/:tmdbId/play", MoviePlaySelector("play", btService))
		movie.GET("/:tmdbId/forceplay", MoviePlaySelector("forceplay", btService))
		movie.GET("/:tmdbId/watchlist/add", AddMovieToWatchlist)
		movie.GET("/:tmdbId/watchlist/remove", RemoveMovieFromWatchlist)
		movie.GET("/:tmdbId/collection/add", AddMovieToCollection)
		movie.GET("/:tmdbId/collection/remove", RemoveMovieFromCollection)
	}

	shows := r.Group("/shows")
	{
		shows.GET("/", TVIndex)
		shows.GET("/search", SearchShows)
		shows.GET("/popular", PopularShows)
		shows.GET("/popular/:genre", PopularShows)
		shows.GET("/recent/shows", RecentShows)
		shows.GET("/recent/shows/:genre", RecentShows)
		shows.GET("/recent/episodes", RecentEpisodes)
		shows.GET("/recent/episodes/:genre", RecentEpisodes)
		shows.GET("/top", TopRatedShows)
		shows.GET("/mostvoted", TVMostVoted)
		shows.GET("/genres", TVGenres)

		trakt := shows.Group("/trakt")
		{
			trakt.GET("/watchlist", WatchlistShows)
			trakt.GET("/collection", CollectionShows)
			trakt.GET("/popular", TraktPopularShows)
			trakt.GET("/trending", TraktTrendingShows)
			trakt.GET("/played", TraktMostPlayedShows)
			trakt.GET("/watched", TraktMostWatchedShows)
			trakt.GET("/collected", TraktMostCollectedShows)
			trakt.GET("/anticipated", TraktMostAnticipatedShows)
			trakt.GET("/progress", TraktProgressShows)
			trakt.GET("/history", TraktHistoryShows)

			lists := trakt.Group("/lists")
			{
				lists.GET("/", TVTraktLists)
				lists.GET("/id/:listId", UserlistShows)
			}

			calendars := trakt.Group("/calendars")
			{
				calendars.GET("/", CalendarShows)
				calendars.GET("/shows", TraktMyShows)
				calendars.GET("/newshows", TraktMyNewShows)
				calendars.GET("/premieres", TraktMyPremieres)
				calendars.GET("/allshows", TraktAllShows)
				calendars.GET("/allnewshows", TraktAllNewShows)
				calendars.GET("/allpremieres", TraktAllPremieres)
			}
		}
	}
	show := r.Group("/show")
	{
		show.GET("/:showId/seasons", ShowSeasons)
		show.GET("/:showId/season/:season/links", ShowSeasonLinks(btService))
		show.GET("/:showId/season/:season/play", ShowSeasonPlay(btService))
		show.GET("/:showId/season/:season/episodes", ShowEpisodes)
		show.GET("/:showId/season/:season/episode/:episode/infolabels", InfoLabelsEpisode(btService))
		show.GET("/:showId/season/:season/episode/:episode/play", ShowEpisodePlaySelector("play", btService))
		show.GET("/:showId/season/:season/episode/:episode/forceplay", ShowEpisodePlaySelector("forceplay", btService))
		show.GET("/:showId/season/:season/episode/:episode/links", ShowEpisodePlaySelector("links", btService))
		show.GET("/:showId/season/:season/episode/:episode/forcelinks", ShowEpisodePlaySelector("forcelinks", btService))
		show.GET("/:showId/watchlist/add", AddShowToWatchlist)
		show.GET("/:showId/watchlist/remove", RemoveShowFromWatchlist)
		show.GET("/:showId/collection/add", AddShowToCollection)
		show.GET("/:showId/collection/remove", RemoveShowFromCollection)
	}
	// TODO
	// episode := r.Group("/episode")
	// {
	// 	episode.GET("/:episodeId/watchlist/add", AddEpisodeToWatchlist)
	// }

	library := r.Group("/library")
	{
		library.GET("/movie/add/:tmdbId", AddMovie)
		library.GET("/movie/remove/:tmdbId", RemoveMovie)
		library.GET("/movie/list/add/:listId", AddMoviesList)
		library.GET("/movie/play/:tmdbId", PlayMovie(btService))
		library.GET("/show/add/:tmdbId", AddShow)
		library.GET("/show/remove/:tmdbId", RemoveShow)
		library.GET("/show/list/add/:listId", AddShowsList)
		library.GET("/show/play/:showId/:season/:episode", PlayShow(btService))

		library.GET("/update", UpdateLibrary)

		// DEPRECATED
		library.GET("/play/movie/:tmdbId", PlayMovie(btService))
		library.GET("/play/show/:showId/season/:season/episode/:episode", PlayShow(btService))
	}

	context := r.Group("/context")
	{
		context.GET("/:media/:kodiID/play", ContextPlaySelector(btService))
	}

	provider := r.Group("/provider")
	{
		provider.GET("/", ProviderList)
		provider.GET("/:provider/check", ProviderCheck)
		provider.GET("/:provider/enable", ProviderEnable)
		provider.GET("/:provider/disable", ProviderDisable)
		provider.GET("/:provider/failure", ProviderFailure)
		provider.GET("/:provider/settings", ProviderSettings)

		provider.GET("/:provider/movie/:tmdbId", ProviderGetMovie)
		provider.GET("/:provider/show/:showId/season/:season/episode/:episode", ProviderGetEpisode)
	}

	allproviders := r.Group("/providers")
	{
		allproviders.GET("/enable", ProvidersEnableAll)
		allproviders.GET("/disable", ProvidersDisableAll)
	}

	repo := r.Group("/repository")
	{
		repo.GET("/:user/:repository/*filepath", repository.GetAddonFiles)
		repo.HEAD("/:user/:repository/*filepath", repository.GetAddonFilesHead)
	}

	trakt := r.Group("/trakt")
	{
		trakt.GET("/authorize", AuthorizeTrakt)
		trakt.GET("/update", UpdateTrakt)
	}

	r.GET("/migrate/:plugin", MigratePlugin)

	r.GET("/setviewmode/:content_type", SetViewMode)

	r.GET("/subtitles", SubtitlesIndex)
	r.GET("/subtitle/:id", SubtitleGet)

	r.GET("/play", Play(btService))
	r.GET("/playuri", PlayURI(btService))

	r.POST("/callbacks/:cid", providers.CallbackHandler)

	// r.GET("/notification", Notification(btService))

	r.GET("/versions", Versions(btService))

	cmd := r.Group("/cmd")
	{
		cmd.GET("/clear_cache", ClearCache)
		cmd.GET("/clear_cache_key/:key", ClearCache)
		cmd.GET("/clear_page_cache", ClearPageCache)
		cmd.GET("/clear_trakt_cache", ClearTraktCache)
		cmd.GET("/clear_tmdb_cache", ClearTmdbCache)
		cmd.GET("/reset_clearances", ResetClearances)
	}

	return r
}
