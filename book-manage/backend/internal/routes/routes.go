package routes

import (
	"github.com/nextsurfer/book-manage-api/internal/app/handler"
)

// add routers
func addRoutes() {
	// deposit address router
	bookHandler := handler.NewBookHandler()
	routerBook := &RouterGroup{
		path: "/book",
		routes: []*RouterOption{
			{
				requestMethod: RequestMethodPost,
				path:          "/add",
				handlerFunc:   bookHandler.AddBook,
			},
			{
				requestMethod: RequestMethodGet,
				path:          "/allbooks",
				handlerFunc:   bookHandler.GetAllBooks,
			},
			{
				requestMethod: RequestMethodGet,
				path:          "/csv",
				handlerFunc:   bookHandler.GetCSV,
			},
			{
				requestMethod: RequestMethodGet,
				path:          "/bundle",
				handlerFunc:   bookHandler.GetBundle,
			},
			{
				requestMethod: RequestMethodPost,
				path:          "/update",
				handlerFunc:   bookHandler.UpdateBook,
			},
			{
				requestMethod: RequestMethodGet,
				path:          "/get",
				handlerFunc:   bookHandler.GetBook,
			},
			{
				requestMethod: RequestMethodGet,
				path:          "/preview",
				handlerFunc:   bookHandler.SearchBookItem,
			},
			{
				requestMethod: RequestMethodGet,
				path:          "/tts",
				handlerFunc:   bookHandler.TextToSpeach,
			},
			{
				requestMethod: RequestMethodGet,
				path:          "/logs",
				handlerFunc:   bookHandler.GetUploadingLog,
			},
			{
				requestMethod: RequestMethodPost,
				path:          "/update_preview",
				handlerFunc:   bookHandler.UpdatePreview,
			},
			{
				requestMethod: RequestMethodGet,
				path:          "/example_position",
				handlerFunc:   bookHandler.GetExamplePosition,
			},
			{
				requestMethod: RequestMethodPost,
				path:          "/delete_preview",
				handlerFunc:   bookHandler.DeletePreview,
			},
			{
				requestMethod: RequestMethodGet,
				path:          "/search",
				handlerFunc:   bookHandler.SearchStringPagination,
			},
			{
				requestMethod: RequestMethodGet,
				path:          "/cefr_levels",
				handlerFunc:   bookHandler.GetCefrLevels,
			},
			{
				requestMethod: RequestMethodGet,
				path:          "/next_sort_value",
				handlerFunc:   bookHandler.GetNextSortValue,
			},
			{
				requestMethod: RequestMethodPost,
				path:          "/update_cefr_level",
				handlerFunc:   bookHandler.UpdateCefrLevel,
			},
			{
				requestMethod: RequestMethodGet,
				path:          "/definition_info",
				handlerFunc:   bookHandler.GetDefinitionInfo,
			},
			{
				requestMethod: RequestMethodPost,
				path:          "/new_definition",
				handlerFunc:   bookHandler.NewDefinition,
			},
			{
				requestMethod: RequestMethodGet,
				path:          "/list_definition",
				handlerFunc:   bookHandler.ListDefinition,
			},
		},
	}

	operateLogHandler := handler.NewOperateLogHandler()
	routerOperateLog := &RouterGroup{
		path: "/operate_log",
		routes: []*RouterOption{
			{
				requestMethod: RequestMethodGet,
				path:          "/pagination",
				handlerFunc:   operateLogHandler.GetPagination,
			},
			{
				requestMethod: RequestMethodGet,
				path:          "/workingEthics",
				handlerFunc:   operateLogHandler.GetWorkingEthics,
			},
			{
				requestMethod: RequestMethodGet,
				path:          "/preview_latest_logs",
				handlerFunc:   operateLogHandler.GetPreviewLatestLogs,
			},
		},
	}

	systemHandler := handler.NewSystemHandler()
	routerSystem := &RouterGroup{
		path: "/system",
		routes: []*RouterOption{
			{
				requestMethod: RequestMethodGet,
				path:          "/check_admin",
				handlerFunc:   systemHandler.CheckAdmin,
			},
			{
				requestMethod: RequestMethodGet,
				path:          "/staffs",
				handlerFunc:   systemHandler.ListStaff,
			},
		},
	}

	backupHandler := handler.NewBackupHandler()
	routerBackupLog := &RouterGroup{
		path: "/backup",
		routes: []*RouterOption{
			{
				requestMethod: RequestMethodGet,
				path:          "/all",
				handlerFunc:   backupHandler.ListBackups,
			},
			{
				requestMethod: RequestMethodPost,
				path:          "/make",
				handlerFunc:   backupHandler.MakeBackup,
			},
			{
				requestMethod: RequestMethodPost,
				path:          "/regain",
				handlerFunc:   backupHandler.RegainBackup,
			},
			{
				requestMethod: RequestMethodGet,
				path:          "/logs",
				handlerFunc:   backupHandler.CheckRegainingLog,
			},
			{
				requestMethod: RequestMethodGet,
				path:          "/cron_status",
				handlerFunc:   backupHandler.GetCronStatus,
			},
			{
				requestMethod: RequestMethodPost,
				path:          "/update_cron_setting",
				handlerFunc:   backupHandler.UpdateCronSetting,
			},
		},
	}

	routerGrpList = append(routerGrpList,
		routerBook,
		routerOperateLog,
		routerSystem,
		routerBackupLog,
	)
}
