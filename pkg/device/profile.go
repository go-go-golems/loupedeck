package device

import "fmt"

// DeviceProfile describes the hardware capabilities of a specific Loupedeck model.
type DeviceProfile struct {
	ProductID string
	Name      string
	Displays  []DisplaySpec
}

// DisplaySpec describes one physical display region.
type DisplaySpec struct {
	Name      string
	ID        byte
	Width     int
	Height    int
	OffsetX   int
	OffsetY   int
	BigEndian bool
}

var deviceProfiles = map[string]DeviceProfile{
	"0003": {
		ProductID: "0003",
		Name:      "Loupedeck CT v1",
		Displays: []DisplaySpec{
			{Name: "left", ID: 'L', Width: 60, Height: 270, OffsetX: 0, OffsetY: 0, BigEndian: false},
			{Name: "main", ID: 'A', Width: 360, Height: 270, OffsetX: 60, OffsetY: 0, BigEndian: false},
			{Name: "right", ID: 'R', Width: 60, Height: 270, OffsetX: 420, OffsetY: 0, BigEndian: false},
			{Name: "dial", ID: 'W', Width: 240, Height: 240, OffsetX: 0, OffsetY: 0, BigEndian: true},
		},
	},
	"0007": {
		ProductID: "0007",
		Name:      "Loupedeck CT v2",
		Displays: []DisplaySpec{
			{Name: "left", ID: 'M', Width: 60, Height: 270, OffsetX: 0, OffsetY: 0, BigEndian: false},
			{Name: "main", ID: 'M', Width: 360, Height: 270, OffsetX: 60, OffsetY: 0, BigEndian: false},
			{Name: "right", ID: 'M', Width: 60, Height: 270, OffsetX: 420, OffsetY: 0, BigEndian: false},
			{Name: "all", ID: 'M', Width: 480, Height: 270, OffsetX: 0, OffsetY: 0, BigEndian: false},
			{Name: "dial", ID: 'W', Width: 240, Height: 240, OffsetX: 0, OffsetY: 0, BigEndian: true},
		},
	},
	"0004": {
		ProductID: "0004",
		Name:      "Loupedeck Live",
		Displays: []DisplaySpec{
			{Name: "left", ID: 'L', Width: 60, Height: 270, OffsetX: 0, OffsetY: 0, BigEndian: false},
			{Name: "main", ID: 'A', Width: 360, Height: 270, OffsetX: 0, OffsetY: 0, BigEndian: false},
			{Name: "right", ID: 'R', Width: 60, Height: 270, OffsetX: 0, OffsetY: 0, BigEndian: false},
		},
	},
	"0006": {
		ProductID: "0006",
		Name:      "Loupedeck Live S",
		Displays: []DisplaySpec{
			{Name: "left", ID: 'M', Width: 60, Height: 270, OffsetX: 0, OffsetY: 0, BigEndian: false},
			{Name: "main", ID: 'M', Width: 360, Height: 270, OffsetX: 60, OffsetY: 0, BigEndian: false},
			{Name: "right", ID: 'M', Width: 60, Height: 270, OffsetX: 420, OffsetY: 0, BigEndian: false},
			{Name: "all", ID: 'M', Width: 480, Height: 270, OffsetX: 0, OffsetY: 0, BigEndian: false},
		},
	},
	"0d06": {
		ProductID: "0d06",
		Name:      "Razer Stream Controller",
		Displays: []DisplaySpec{
			{Name: "left", ID: 'M', Width: 60, Height: 270, OffsetX: 0, OffsetY: 0, BigEndian: false},
			{Name: "main", ID: 'M', Width: 360, Height: 270, OffsetX: 60, OffsetY: 0, BigEndian: false},
			{Name: "right", ID: 'M', Width: 60, Height: 270, OffsetX: 420, OffsetY: 0, BigEndian: false},
			{Name: "all", ID: 'M', Width: 480, Height: 270, OffsetX: 0, OffsetY: 0, BigEndian: false},
		},
	},
}

func resolveProfile(product string) (DeviceProfile, error) {
	profile, ok := deviceProfiles[product]
	if !ok {
		return DeviceProfile{}, fmt.Errorf("unknown device product ID: %q", product)
	}
	return profile, nil
}

func (l *Loupedeck) applyProfile(profile DeviceProfile) {
	l.Model = profile.Name
	if l.displays == nil {
		l.displays = map[string]*Display{}
	} else {
		for name := range l.displays {
			delete(l.displays, name)
		}
	}
	for _, spec := range profile.Displays {
		l.addDisplay(spec.Name, spec.ID, spec.Width, spec.Height, spec.OffsetX, spec.OffsetY, spec.BigEndian)
	}
}
