// Пакет shortycheck реализует кастомный статический анализатор.
// Включает в себя набор анализаторов из go/analysis:
//
//	assign       - обнаруживает бесполезные присваивания.
//	bools        - обнаруживает распространенные ошибки, связанные с логическими операторами.
//	copylock     - проверяет наличие блокировок, ошибочно переданных по значению.
//	errorsas     - проверяет, соответствует ли второй аргумент ошибкам
//	httpresponse - проверяет наличие ошибок с помощью HTTP-ответов.
//	loopclosure  - проверяет наличие ссылок на переменные окружающего цикла из вложенных функций.
//	lostcancel   - проверяет, не удалось ли вызвать функцию отмены контекста.
//	printf       - проверяет согласованность строк и аргументов формата Printf.
//	tests        - проверяет распространенное ошибочное использование тестов и примеров.
//
// Публичные анализаторы:
//
//	sqlrows      - проверяет наличие ошибок при работе с sql.Rows.
//	signature    - ищет функции с низкой читаемостью.
//
// Также включает (опционально) набор анализаторов из staticcheck (передается в виде аргумента staticcheckAnalyzers при запуске)
// и кастомный анализатор, который проверяет наличие вызовов `os.Exit` в функции main.
package shortycheck

import (
	"strings"

	"github.com/gostaticanalysis/signature"
	"github.com/gostaticanalysis/sqlrows/passes/sqlrows"
	"golang.org/x/exp/slices"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/tests"
	"honnef.co/go/tools/staticcheck"
)

// RunMultichecker запускает кастомный мультичекер.
func RunMultichecker(staticcheckAnalyzers []string) {
	// https://github.com/gostaticanalysis
	mychecks := []*analysis.Analyzer{
		assign.Analyzer,       // обнаруживает бесполезные присваивания.
		bools.Analyzer,        // обнаруживает распространенные ошибки, связанные с логическими операторами.
		copylock.Analyzer,     // проверяет наличие блокировок, ошибочно переданных по значению.
		errorsas.Analyzer,     // проверяет, соответствует ли второй аргумент ошибкам
		httpresponse.Analyzer, // проверяет наличие ошибок с помощью HTTP-ответов.
		loopclosure.Analyzer,  // проверяет наличие ссылок на переменные окружающего цикла из вложенных функций.
		lostcancel.Analyzer,   // проверяет, не удалось ли вызвать функцию отмены контекста.
		printf.Analyzer,       // проверяет согласованность строк и аргументов формата Printf.
		tests.Analyzer,        // проверяет распространенное ошибочное использование тестов и примеров.
		sqlrows.Analyzer,      // проверяет наличие ошибок при работе с sql.Rows.
		signature.Analyzer,    // ищет функции с низкой читаемостью.
		MainExitAnalyzer,      // который проверяет наличие вызовов os.Exit в функции main.
	}

	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") ||
			slices.Contains(staticcheckAnalyzers, v.Analyzer.Name) {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	multichecker.Main(
		mychecks...,
	)

}
