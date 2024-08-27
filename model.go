package glitch

type Model struct {
	mesh     *Mesh
	material Material
}

func NewModel(mesh *Mesh, material Material) *Model {
	return &Model{
		mesh:     mesh,
		material: material,
	}
}

func (m *Model) Draw(target BatchTarget, matrix Mat4) {
	m.DrawColorMask(target, matrix, White)
}
func (m *Model) DrawColorMask(target BatchTarget, matrix Mat4, mask RGBA) {
	target.Add(m.mesh, matrix.gl(), mask, m.material, false)
}
