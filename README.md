# Vim Master

Aprende Vim jugando, directamente en tu terminal.

Vim Master es un juego de terminal (TUI) con un editor que emula Vim. Cada ejercicio te pide transformar un texto o llevar el cursor a un lugar usando comandos reales de Vim. El juego comprueba el resultado y te da puntos según tu velocidad, cuántas teclas usaste (estilo VimGolf) y tus combos.

- **12 módulos y 105 ejercicios**, de `hjkl` hasta macros.
- **4 modos de juego**: Lecciones, Contrarreloj, Golf y Examen final.
- **Puntaje con rango** (S, A, B, C, D) y ranking local de los 10 mejores por modo.
- Progreso guardado: cada módulo se desbloquea al aprobar el anterior.
- Un único binario, sin dependencias, para Linux, macOS y Windows.

## Instalación

### Con `go install`

Requiere [Go](https://go.dev/dl/) 1.27 o superior.

```bash
go install github.com/Rickemtz/vim-master/cmd/vim-master@latest
vim-master
```

El binario queda en `$(go env GOPATH)/bin`, que debe estar en tu `PATH`.

### Desde el código fuente

```bash
git clone https://github.com/Rickemtz/vim-master.git
cd vim-master
make build
./bin/vim-master
```

Para jugar sin compilar: `make run`.

### Binarios precompilados

Si hay una versión publicada, descarga el archivo para tu sistema desde la página de [Releases](https://github.com/Rickemtz/vim-master/releases), descomprímelo y ejecuta `vim-master`.

## Cómo se juega

Abre `vim-master` y elige un modo en el menú con `↑/↓` (o `j/k`) y `Enter`.

La pantalla de juego muestra, de arriba abajo:

1. **Marcador**: ejercicio actual, tiempo restante, puntos y combo.
2. **Instrucciones** del ejercicio.
3. **Editor**, donde escribes comandos de Vim.
4. **Objetivo**: el texto esperado, con lo correcto en verde y lo pendiente en rojo.
5. **Barra de estado** al estilo de Vim: modo actual y teclas pendientes (por ejemplo `d2` mientras escribes `d2w`).

El juego usa las teclas de función para no interferir con ningún comando de Vim:

| Tecla | Acción |
|-------|--------|
| `F1`  | Pedir una pista (resta puntos y rompe el combo) |
| `F2`  | Reiniciar el ejercicio |
| `F10` | Volver al menú |

La terminal debe medir al menos **80×24**.

## Modos de juego

| Modo | Descripción |
|------|-------------|
| **Lecciones** | Recorre los módulos en orden, con instrucciones y pistas. Cada ejercicio tiene su propio límite de tiempo. Aprobar un módulo con rango C o mejor desbloquea el siguiente. |
| **Contrarreloj** | Elige 60, 120 o 300 segundos y completa todos los ejercicios que puedas. Salen al azar de los módulos que ya desbloqueaste. |
| **Golf** | Sin límite de tiempo: gana quien usa menos teclas. |
| **Examen final** | 20 ejercicios mezclados de todos los módulos, sin pistas. Da tu rango oficial. |

En el menú, **Ranking** muestra los 10 mejores puntajes de cada modo.

## Módulos

| # | Módulo | Comandos |
|---|--------|----------|
| 1 | Supervivencia | `h j k l`, `i a I A o O`, `Esc`, `x` |
| 2 | Palabras y líneas | `w b e W B E 0 ^ $` |
| 3 | Saltos | `gg G {n}G f F t T ; , %` |
| 4 | Borrar y cambiar | `d c x r s D C dd cc` con movimientos |
| 5 | Deshacer y repetir | `u Ctrl-r .` |
| 6 | Copiar y pegar | `y yy p P`, registros |
| 7 | Text objects | `iw aw i" a" i( a( ip ap`… |
| 8 | Búsqueda | `/ ? n N * #` |
| 9 | Modo visual | `v V Ctrl-v` |
| 10 | Combinaciones | counts, `> < ~ J`, operador + text object |
| 11 | Sustitución | `:s :%s :g` |
| 12 | Macros | `q @ @@` |

## Puntuación

Cada ejercicio vale `100 × dificultad` puntos base, más dos bonos:

- **Tiempo**: hasta +50% del valor base si terminas rápido.
- **Eficiencia**: hasta +50% del valor base si usas tan pocas teclas como el *par* del ejercicio.

Restan puntos las pistas (−15% del valor base cada una), las flechas en los ejercicios que las prohíben y los comandos no permitidos. Si se acaba el tiempo, el ejercicio vale 0.

**Combo**: cada ejercicio resuelto sin pistas y con pocas teclas (hasta 20% por encima del par) suma ×0.1 al multiplicador, hasta ×2.0. Pedir una pista o quedarte sin tiempo lo reinicia.

**Rango final**: S ≥ 95% · A ≥ 85% · B ≥ 70% · C ≥ 50% · D < 50% del máximo posible sin combos.

## Datos guardados

El progreso y los récords se guardan en el directorio de configuración del sistema, en la carpeta `vim-master/`: por ejemplo `~/.config/vim-master/` en Linux. Allí están `progress.json` y `scores.json`. Para empezar desde cero, borra esa carpeta.

## Desarrollo

```bash
make run              # ejecuta el juego
make build            # compila en ./bin/vim-master
make test             # corre los tests
make lint             # go vet + gofmt
make validate         # resuelve cada ejercicio con su solución y verifica el resultado
make release-snapshot # compila para todas las plataformas con goreleaser, sin publicar
```

Estructura:

- `internal/engine`: el emulador de Vim, sin dependencias de la interfaz.
- `internal/scoring`: puntos, combos y rangos.
- `internal/ui`: pantallas con [Bubble Tea](https://github.com/charmbracelet/bubbletea). Todos los textos visibles están en `internal/ui/strings.go`.
- `exercises/`: un archivo YAML por ejercicio. `claude.md` describe el formato.

### Agregar un ejercicio

Crea un YAML en la carpeta del módulo, por ejemplo `exercises/02-palabras/011-mi-ejercicio.yaml`, siguiendo el formato de los demás. Incluye `solution` con las teclas que lo resuelven y comprueba que `make validate` pase.
