# Fixtures

Real lifebar definition files, committed unmodified as reference fixtures
(roadmap decision `044`: reference content is committed as downloaded,
licence notwithstanding). Plain text only: no sprites, fonts or sounds.

| File | Source | Origin |
|---|---|---|
| `ikemen-default-fight.def` | `~/.cache/kakutou-fidelity/work/ikemen/data/fight.def` | The default fight screen of an Ikemen GO release build (motif IKEMEN1, 1280x720). The build's commit is unknown. |
| `vhd-ikemen-go-fixture.def` | `lifebar-viewer-web/src/lifebar/testdata/` | V-HD Bars community pack, trimmed (see below). |
| `gms-ikemen-go-fixture.def` | same | GMS community pack, trimmed. Has `[Guardbar]`, `[Stunbar]` and `[Simul_3P Lifebar]`. |
| `mfj2-classic-mugen-fixture.def` | same | MFJ2 community pack, trimmed. MUGEN section naming, CRLF and LF mixed, comments already corrupted to U+FFFD. |
| `truncated-real-fixture.def` | same | Excerpt of the GMS fixture with a section header's closing `]` cut off. |
| `game-default-fight.def` | `game/native/testdata/content/lifebars/default/fight.def` | OpenKakutou's own test lifebar (solid bars); original work, MIT. |

The four fixtures from `lifebar-viewer-web` were trimmed from lifebar packs of
a personal reference collection, informally distributed as freeware in the
MUGEN/Ikemen GO community with no formal licence. Their provenance is
described in `lifebar-viewer-web/src/lifebar/testdata/README.md`; the files
are copied here unchanged.
