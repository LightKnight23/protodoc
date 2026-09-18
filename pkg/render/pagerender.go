// Open-and-render-a-page memory/octet budgets (T-0262, T-0263). Opening a
// document and rendering any single page is bounded: peak working memory stays
// within the page-render budget (NFR-017) and the octets read to render one
// page stay within the per-page read budget plus the page's resources (that
// octet-read budget's own requirement is exercised by T-0263). The render path
// here is a self-contained page render (decode the page's materialised objects
// at their extents into a page raster) that a benchmark can measure.
package render

// PageRenderBudgetBytes is the NFR-017 peak-memory ceiling for opening and
// rendering any single page (<= 200 MiB).
const PageRenderBudgetBytes = 200 * 1024 * 1024

// PageOctetReadBudgetBytes is the base per-page octet-read budget (<= 8 MiB)
// before the page's own resource octets are added (exercised by T-0263).
const PageOctetReadBudgetBytes = 8 * 1024 * 1024

// PageToRender is one page's render inputs: the page geometry and the embedded
// objects placed on it (each with a materialised representation).
type PageToRender struct {
	Width, Height int
	Objects       []EmbeddedObject
}

// RenderPage renders a single page into an RGB raster of its geometry by
// compositing each embedded object's materialised representation at its
// declared extent. It allocates one page raster plus each object's decoded
// raster transiently; nothing scales with the whole document. It also reports
// the octets read to render the page (the page's own resource payloads).
func RenderPage(page PageToRender) (raster []byte, octetsRead int, err error) {
	raster = make([]byte, page.Width*page.Height*3)
	for i := range raster {
		raster[i] = backgroundSample
	}
	for _, obj := range page.Objects {
		octetsRead += len(obj.Payload)
		ro, rerr := RenderMaterialised(obj)
		if rerr != nil {
			return nil, octetsRead, rerr
		}
		compositeInto(raster, page.Width, page.Height, obj.DeclExtent, ro.Raster, ro.Channels)
	}
	return raster, octetsRead, nil
}

// compositeInto copies an object's RGB(A) raster into the page raster at the
// object's extent, clipping to the page bounds.
func compositeInto(page []byte, pw, ph int, ext Extent, src []byte, channels int) {
	for y := 0; y < ext.Height; y++ {
		py := ext.Y + y
		if py < 0 || py >= ph {
			continue
		}
		for x := 0; x < ext.Width; x++ {
			px := ext.X + x
			if px < 0 || px >= pw {
				continue
			}
			si := (y*ext.Width + x) * channels
			di := (py*pw + px) * 3
			if si+2 < len(src) {
				page[di] = src[si]
				page[di+1] = src[si+1]
				page[di+2] = src[si+2]
			}
		}
	}
}
