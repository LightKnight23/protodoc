// Colour: the one normative colour representation (CON-014, plan.md CON-014
// satisfaction row). Two distinct types express two distinct roles (CP-008:
// exactly one representation per capability) -- StoredColour is the
// persisted, gamma-encoded component form; CompositingColour is the
// linear-light space in which blending arithmetic is performed. No third
// type, and no branch selecting an alternate colour space, exists anywhere
// in this package.
package container

// ColourProfileID identifies the colour representation a document is
// authored against (contracts/container.abnf fm-colour-profile-id
// NORMATIVE comment). CON-014 requires the format to define exactly one
// such representation; ColourProfileSRGBD65 is that sole registry-issued
// id for this format version.
type ColourProfileID uint16

// ColourProfileSRGBD65 is CON-014's one normative colour representation:
// sRGB primaries, the D65 white point, the sRGB piecewise transfer
// function, StoredColour's u16 per-channel stored component width, and
// CompositingColour's linear-light 32-bit fixed-point compositing space.
// No second id is defined for this format version.
const ColourProfileSRGBD65 ColourProfileID = 1

// ColourComponentBits is StoredColour's per-channel stored component width
// CON-014 fixes: u16.
const ColourComponentBits = 16

// CompositingFixedPointBits is CompositingColour's per-channel width
// CON-014 fixes: 32-bit fixed-point, linear light.
const CompositingFixedPointBits = 32

// StoredColour is CON-014's persisted stored-colour representation: four
// u16 components (red, green, blue, alpha), gamma-encoded against sRGB
// primaries and the D65 white point. This is the sole representation of a
// persisted colour value in this package (CP-008): no second, alternate
// stored-colour type exists.
type StoredColour struct {
	R, G, B, A uint16
}

// CompositingColour is CON-014's linear-light compositing-space
// representation: four 32-bit fixed-point components, the space in which
// all blending/compositing arithmetic is performed. It is distinct from
// StoredColour (gamma-encoded, wire-persisted) -- exactly one type exists
// for each of the two roles (CP-008), with no alternate colour space
// reachable from either.
type CompositingColour struct {
	R, G, B, A uint32
}
