# CLAUDE.md — Vim Master

Guía de trabajo para Claude Code en este repositorio. Léela completa antes de cada tarea.

## Qué es este proyecto

**Vim Master** es una aplicación de terminal (TUI) para aprender Vim jugando. Combina tres ideas:

1. **Lecciones progresivas** estilo "Vim Master" / vimtutor: módulos que enseñan comandos de menor a mayor dificultad.
2. **Contrarreloj**: cada ejercicio tiene un límite de tiempo; terminar rápido da bonus.
3. **Sistema de puntos**: cada ejercicio otorga puntos según tiempo, eficiencia de teclas (estilo VimGolf) y combos. Al final de cada sesión se muestra un **puntaje final con rango** (S, A, B, C, D) y se guarda en un ranking local.

El usuario escribe comandos Vim reales dentro de un editor emulado. La app valida el resultado (contenido del buffer y/o posición del cursor) contra el objetivo del ejercicio.

## Stack

- **Lenguaje:** Go (≥ 1.22). Se eligió por generar un único binario multiplataforma.
- **TUI:** [Bubble Tea](https://github.com/charmbracelet/bubbletea) (arquitectura Elm: Model / Update / View).
- **Estilos:** [Lip Gloss](https://github.com/charmbracelet/lipgloss).
- **Componentes:** [Bubbles](https://github.com/charmbracelet/bubbles) (listas, timer, progress).
- **Ejercicios:** archivos YAML embebidos con `go:embed` (`gopkg.in/yaml.v3`).
- **Persistencia:** JSON en `~/.config/vim-master/` (progreso y récords).
- **Tests:** `testing` estándar + tablas de casos. Sin frameworks extra.

No agregar dependencias nuevas sin preguntar primero.

## Comandos

```bash
make run              # go run ./cmd/vim-master
make build            # binario en ./bin/vim-master, con versión/commit/fecha inyectados
make test             # go test ./...
make lint             # go vet ./... && gofmt -l .
make validate         # valida todos los ejercicios YAML (go run ./cmd/validate)
make release-snapshot # build multiplataforma local con goreleaser, sin publicar nada
```

`vim-master --version` muestra versión/commit/fecha de build. El release real (tags + publicar en GitHub) se hace con `goreleaser release` (ver `.goreleaser.yaml`); necesita un repo git con remote y un tag.

Antes de dar una tarea por terminada: `make test` y `make lint` deben pasar.

## Estructura del proyecto

```
vim-master/
├── cmd/
│   ├── vim-master/       # main: arranca la TUI
│   └── validate/         # valida ejercicios: resuelve cada uno con su "solution" y comprueba el target
├── internal/
│   ├── engine/           # emulador de Vim (SIN dependencias de UI)
│   │   ├── buffer.go     # líneas de texto, inserción/borrado
│   │   ├── cursor.go     # posición, clamping, columna deseada
│   │   ├── modes.go      # Normal, Insert, Visual, VisualLine, VisualBlock, Command
│   │   ├── parser.go     # parser de comandos: [count][operador][count][movimiento|text-object]
│   │   ├── motions.go    # h j k l w b e 0 ^ $ gg G f t % ...
│   │   ├── operators.go  # d c y > < ~ ...
│   │   ├── textobjects.go# iw aw i" a" i( a( ip ...
│   │   ├── registers.go  # registro sin nombre, "0, "a-"z
│   │   ├── undo.go       # pila de undo/redo
│   │   ├── search.go     # / ? n N * #
│   │   ├── excmd.go      # :w :q :s :%s :g (subconjunto)
│   │   └── engine.go     # API pública: Feed(key) -> estado
│   ├── exercise/         # carga y validación de ejercicios YAML
│   ├── scoring/          # cálculo de puntos, combos y rangos (funciones puras)
│   ├── session/          # orquesta una partida: lista de ejercicios, timer, acumulado
│   ├── storage/          # lectura/escritura de progreso y récords
│   └── ui/               # Bubble Tea: pantallas y componentes
│       ├── app.go        # modelo raíz, enrutado entre pantallas, persistencia al terminar una sesión
│       ├── menu.go
│       ├── duration.go   # elegir duración de Contrarreloj
│       ├── lesson.go     # pantalla de juego (editor + HUD); un solo modelo para Lecciones/Contrarreloj/Golf/Examen
│       ├── results.go    # resultado final con rango
│       ├── leaderboard.go# top 10 por modo
│       ├── keys.go       # traduce teclas de Bubble Tea a engine.Key
│       └── styles.go
├── exercises/            # YAML por módulo: 01-supervivencia/, 02-palabras/, ..., 12-macros/
│   └── embed.go          # go:embed de los YAML
├── .goreleaser.yaml       # build multiplataforma para releases
├── Makefile
└── CLAUDE.md
```

### Regla de arquitectura clave

`internal/engine` y `internal/scoring` **no importan nada de `ui`** ni de Bubble Tea. Son lógica pura y 100% testeable. La UI solo traduce `tea.KeyMsg` a teclas del engine y dibuja su estado.

## Motor de Vim (`internal/engine`)

### API

```go
type Engine struct { /* ... */ }

func New(text string, cursor Pos) *Engine
func (e *Engine) Feed(key Key) Event   // procesa una tecla
func (e *Engine) Text() string
func (e *Engine) Cursor() Pos
func (e *Engine) Mode() Mode
func (e *Engine) PendingKeys() string  // p.ej. "d2" mientras se escribe "d2w"
func (e *Engine) Stats() Stats         // teclas totales, teclas de flecha, comandos usados
```

`Key` representa teclas normales y especiales (`Esc`, `Enter`, `Backspace`, `Ctrl-r`, `Ctrl-v`, flechas).

### Comportamiento

- Emular el comportamiento de **Vim por defecto** (no Neovim con configuración), con `nocompatible`.
- Ante dudas sobre la semántica exacta de un comando, seguir `:help` de Vim y documentar el caso en un test.
- Comandos no soportados: no fallar; emitir un `Event` de tipo `Unsupported` para que la UI muestre "comando aún no disponible".
- Registrar en `Stats` cada tecla y cada comando completo (para detectar uso de flechas y comandos permitidos).

### Alcance por fases

Implementar en este orden. No adelantar fases sin pedirlo.

| Fase | Contenido |
|------|-----------|
| E1 | Buffer, cursor, modos Normal/Insert, `h j k l`, `i a I A o O`, `Esc`, `x`, `:w :q` (simulados) |
| E2 | `w b e W B E 0 ^ $ gg G {n}G`, counts |
| E3 | Operadores `d c y` + movimientos, `dd cc yy D C`, `p P`, `r s`, `u Ctrl-r` |
| E4 | `f F t T ; ,`, `%`, `.` (repetir) |
| E5 | Text objects: `iw aw iW aW i" a" i' a' i( a( i[ a[ i{ a{ ip ap it at` |
| E6 | Búsqueda `/ ? n N * #` |
| E7 | Visual `v V Ctrl-v` con operadores |
| E8 | `> < ~ J`, registros con nombre `"a` |
| E9 | Ex: `:s`, `:%s`, rangos simples, `:g` básico |
| E10 | Macros `q{reg}` y `@{reg}`, `@@` |

## Formato de ejercicios

Un archivo YAML por ejercicio en `exercises/<NN-modulo>/`.

```yaml
id: palabras-003
module: 2
title: "Salta de palabra en palabra"
difficulty: 2              # 1-5, afecta los puntos base
type: edit                 # edit | cursor | both
instructions: |
  Borra la palabra "rojo" usando movimientos de palabra.
initial: |
  El coche rojo es rápido
cursor_start: [0, 0]       # [línea, columna], base 0
target: |
  El coche es rápido
target_cursor: null        # obligatorio si type es cursor o both
time_limit_sec: 30
par_keystrokes: 4          # mínimo razonable de teclas (estilo VimGolf)
solution: "wwdw"           # usado por `make validate`; debe alcanzar el target en <= par
allowed: [w, b, e, d, x]   # opcional: comandos permitidos; vacío = todos
forbid_arrows: true
hints:
  - "w avanza al inicio de la siguiente palabra"
  - "dw borra hasta el inicio de la siguiente palabra"
```

Reglas:
- `solution` es obligatoria y debe pasar `make validate`.
- `par_keystrokes` debe ser igual a la longitud en teclas de `solution` (o menor si existe una más corta conocida).
- Textos en español; ejemplos de código pueden ser en cualquier lenguaje.
- Mínimo 8 ejercicios por módulo antes de considerarlo terminado.

## Currículo (módulos)

1. **Supervivencia**: modos, `i`, `Esc`, `hjkl`, `x`, `:wq`
2. **Palabras y líneas**: `w b e W B E 0 ^ $`
3. **Saltos**: `gg G {n}G f F t T ; , %`
4. **Borrar y cambiar**: `d c x r s D C dd cc` con movimientos
5. **Deshacer y repetir**: `u Ctrl-r .`
6. **Copiar y pegar**: `y yy p P`, registros básicos
7. **Text objects**: `iw aw i" a( ip`, etc.
8. **Búsqueda**: `/ ? n N * #`
9. **Modo visual**: `v V Ctrl-v`
10. **Combinaciones**: counts, `> < ~ J`, operador + text object
11. **Sustitución**: `:s :%s :g`
12. **Macros**: `q @ @@`
13. **Examen final**: ejercicios mixtos de todos los módulos

Un módulo se desbloquea al completar el anterior con rango C o superior.

## Sistema de puntuación (`internal/scoring`)

Funciones puras, con tests para cada fórmula. Todos los valores son `int` redondeados al final.

### Por ejercicio

```
base        = 100 × difficulty

bonus_tiempo    = base × 0.5 × max(0, 1 − t_usado / time_limit)
bonus_eficiencia= base × 0.5 × min(1, par_keystrokes / teclas_usadas)

penalizaciones:
  - pista usada:        −15% de base por pista
  - tecla de flecha:    −5 puntos cada una (si forbid_arrows)
  - comando no permitido: −10 puntos cada uno

subtotal = max(0, base + bonus_tiempo + bonus_eficiencia − penalizaciones)
puntos   = subtotal × multiplicador_combo
```

- Si se agota el tiempo: el ejercicio vale **0** y se rompe el combo. El usuario puede continuar al siguiente.
- Máximo teórico por ejercicio (sin combo): `base × 2`.

### Combo

- Sube +0.1 cuando el ejercicio se completa **sin pistas** y con `teclas_usadas ≤ par × 1.2`.
- Tope: ×2.0.
- Se reinicia a ×1.0 al usar pista, agotar el tiempo o saltar el ejercicio.

### Puntaje final y rango

```
porcentaje = puntaje_total / máximo_teórico_sin_combo × 100
```

| Rango | Porcentaje |
|-------|-----------|
| S | ≥ 95% |
| A | ≥ 85% |
| B | ≥ 70% |
| C | ≥ 50% |
| D | < 50% |

(Con combos se puede superar el 100%; eso sigue siendo S.)

La pantalla final muestra: puntaje total, rango, tiempo total, teclas totales vs par total, precisión (ejercicios completados / total), combo máximo, comandos más usados y si es un nuevo récord.

## Modos de juego

- **Lecciones**: módulo a módulo, con instrucciones y pistas. Guarda progreso.
- **Contrarreloj**: 60/120/300 segundos globales, ejercicios aleatorios de módulos desbloqueados; puntúa cuantos más completes.
- **Golf**: sin límite de tiempo, solo cuenta la eficiencia de teclas (estilo VimGolf).
- **Examen final**: 20 ejercicios mixtos de TODOS los módulos (no solo los desbloqueados, a diferencia de Contrarreloj/Golf), sin pistas, da el rango "oficial".

## Interfaz

Pantalla de juego (orden vertical):

1. **HUD superior**: módulo/ejercicio, temporizador (cambia a amarillo < 30%, rojo < 10%), puntos acumulados, combo.
2. **Instrucciones** del ejercicio.
3. **Editor**: buffer con números de línea, cursor visible (bloque en Normal, barra en Insert), selección resaltada en Visual.
4. **Objetivo**: vista del texto esperado con diferencias resaltadas (verde lo correcto, rojo lo pendiente).
5. **Barra de estado estilo Vim**: modo (`-- INSERT --`), teclas pendientes, línea de comandos `:`.
6. **Ayuda**: `F1` pista, `F2` reiniciar ejercicio, `F10` salir al menú.

Las teclas de control de la app usan `F1–F10` para no chocar con ningún comando de Vim.

Respetar tamaño mínimo de terminal 80×24; si es menor, mostrar aviso en lugar de romper el layout. Colores definidos solo en `ui/styles.go`.

## Persistencia

`~/.config/vim-master/` (usar `os.UserConfigDir()`):
- `progress.json`: módulos desbloqueados, mejor rango por módulo, ejercicios completados.
- `scores.json`: top 10 por modo de juego (nombre, puntaje, rango, fecha).

Escribir en archivo temporal + rename para no corromper datos. Si el JSON está dañado, respaldarlo como `.bak` y empezar de cero.

## Testing

- **Engine**: tests de tabla con `initial`, `cursor`, `keys`, `wantText`, `wantCursor`, `wantMode`. Un archivo `_test.go` por archivo de engine. Cada bug corregido añade un caso.
- **Scoring**: casos límite (tiempo 0, tiempo agotado, teclas < par, combo en tope, muchas penalizaciones que darían negativo).
- **Ejercicios**: `make validate` resuelve cada ejercicio con su `solution` usando el engine real.
- **UI**: tests mínimos de `Update` enviando mensajes; no testear el render píxel a píxel.

## Convenciones de código

- `gofmt` obligatorio; nombres en inglés en el código, textos visibles al usuario en español.
- Todos los textos de UI en `internal/ui/strings.go` (facilita traducir después).
- Errores con contexto: `fmt.Errorf("cargar ejercicio %s: %w", id, err)`.
- Sin estado global; pasar dependencias explícitamente.
- Funciones cortas; si un `switch` de teclas crece mucho, dividir por modo.
- Comentarios solo donde la semántica de Vim no es obvia (ej. por qué `cw` se comporta como `ce`).

## Hoja de ruta

- [x] **Fase 0**: esqueleto del proyecto, Makefile, `go.mod`, menú vacío que abre y cierra
- [x] **Fase 1**: engine E1 + pantalla de juego + 8 ejercicios del módulo 1 + timer
- [x] **Fase 2**: scoring completo + pantalla de resultados + puntaje final con rango
- [x] **Fase 3**: persistencia de progreso y ranking
- [x] **Fase 4**: engine E2–E4 + módulos 2–5
- [x] **Fase 5**: modos Contrarreloj y Golf
- [x] **Fase 6**: engine E5–E7 + módulos 6–9
- [x] **Fase 7**: engine E8–E10 + módulos 10–12
- [x] **Fase 8**: Examen final, pulido visual, release con goreleaser

Marcar las casillas al completar cada fase.

## Cómo trabajar en este repo (instrucciones para Claude)

- Trabajar **una fase o subtarea a la vez**; al terminar, resumir qué cambió y proponer el siguiente paso.
- Antes de implementar un comando de Vim, escribir primero sus tests.
- No modificar fórmulas de puntuación ni el formato YAML sin confirmar con el usuario; si cambian, actualizar este archivo.
- Si una decisión de diseño no está cubierta aquí, preguntar o elegir la opción más simple y anotarla en la sección correspondiente.
- Mantener este CLAUDE.md actualizado cuando cambie la estructura, los comandos o el alcance.
