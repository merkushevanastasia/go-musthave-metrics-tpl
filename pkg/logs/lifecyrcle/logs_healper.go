package lifecyrcle

import "github.com/rs/zerolog"

func LogModuleStart(logger *zerolog.Logger, module string) {
	logger.Info().Str("moduleName", module).Msg("Запуск модуль...")
}

func LogModuleStartedOk(logger *zerolog.Logger, module string) {
	logger.Info().Str("moduleName", module).Msg("Модуль успешно запущен...")
}

func LogModuleDown(logger *zerolog.Logger, module string) {
	logger.Info().Str("moduleName", module).Msg("Модуль остановлен...")
}
