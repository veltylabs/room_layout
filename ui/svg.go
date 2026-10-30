//go:build !wasm

package ui

import (
	"webtyp.com/svg"
	"webtyp.com/svg/sprite"
)

// Icons aloja los íconos del menú de «Espacios»; webtyp/ssr los inyecta una vez en la página y el
// menú los usa con <use href="#…">. Geometría de IcoMoon Free (CC BY 4.0), en 16×16.
type Icons struct{}

func (m *Icons) IconID() string { return ID }

func (m *Icons) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(
		// edificio: el grupo «Espacios»
		sprite.Define(svg.Icon(ID), viewBox16,
			sprite.Path("M0 16h8v-16h-8v16zM5 2h2v2h-2v-2zM5 6h2v2h-2v-2zM5 10h2v2h-2v-2zM1 2h2v2h-2v-2zM1 6h2v2h-2v-2zM1 10h2v2h-2v-2zM9 5h7v1h-7zM9 16h2v-4h3v4h2v-9h-7z")),
		// mapa plegado: el mapa del recinto
		sprite.Define(svg.Icon(iconMap), viewBox16,
			sprite.Path("M0 3l5-2v12l-5 2z"), sprite.Path("M6 0.5l5 3v11.5l-5-2.5z"), sprite.Path("M12 3.5l4-3v12l-4 3z")),
		// puerta: los boxes
		sprite.Define(svg.Icon(iconRoom), viewBox16,
			sprite.Path("M6 8h-5v-2h5v-2l3 3-3 3zM16 0v13l-6 3v-3h-6v-4h1v3h5v-9l4-2h-9v4h-1v-5z")),
		// capas: los pisos
		sprite.Define(svg.Icon(iconFloor), viewBox16,
			sprite.Path("M16 5l-8-4-8 4 8 4 8-4zM8 2.328l5.345 2.672-5.345 2.672-5.345-2.672 5.345-2.672zM14.398 7.199l1.602 0.801-8 4-8-4 1.602-0.801 6.398 3.199zM14.398 10.199l1.602 0.801-8 4-8-4 1.602-0.801 6.398 3.199z")),
		// botiquín: el equipamiento
		sprite.Define(svg.Icon(iconEquipment), viewBox16,
			sprite.Path("M14 4h-3v-2c0-0.55-0.45-1-1-1h-4c-0.55 0-1 0.45-1 1v2h-3c-1.1 0-2 0.9-2 2v8c0 1.1 0.9 2 2 2h12c1.1 0 2-0.9 2-2v-8c0-1.1-0.9-2-2-2zM6 2h4v2h-4v-2zM12 11h-3v3h-2v-3h-3v-2h3v-3h2v3h3v2z")),
	)
}

const viewBox16 = "0 0 16 16"
