package glitch

type Model struct {
	mesh *Mesh
	material Material
}

func NewModel(mesh *Mesh, material Material) *Model {
	return &Model{
		mesh: mesh,
		material: material,
	}
}

func (m *Model) Draw(pass *RenderPass, matrix Mat4) {
	// pass.Add(m.mesh, matrix, RGBA{1.0, 1.0, 1.0, 1.0}, m.material, false)
	m.DrawColorMask(pass, matrix, White)
}
func (m *Model) DrawColorMask(pass *RenderPass, matrix Mat4, mask RGBA) {
	pass.Add(m.mesh, matrix.gl(), mask, m.material, false)
}
