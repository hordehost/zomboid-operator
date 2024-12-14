package settings

import (
	zomboidv1 "github.com/zomboidhost/zomboid-operator/api/v1"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Workshop mod handling", func() {
	When("applying server settings", func() {
		It("should merge WorkshopMods into Mods strings", func() {
			settings := &zomboidv1.ZomboidSettings{
				// Start with some existing mods in the classic format
				Mods: zomboidv1.Mods{
					Mods:          ptr.To("ExistingMod1;ExistingMod2"),
					WorkshopItems: ptr.To("111111;222222"),
				},
				// Add some workshop mods in the structured format
				WorkshopMods: []zomboidv1.WorkshopMod{
					{
						ModID:      ptr.To("NewMod1"),
						WorkshopID: ptr.To("333333"),
					},
					{
						ModID:      ptr.To("NewMod2"),
						WorkshopID: ptr.To("444444"),
					},
				},
			}

			MergeWorkshopMods(settings)

			// Verify the mods were merged correctly
			Expect(*settings.Mods.Mods).To(Equal("ExistingMod1;ExistingMod2;NewMod1;NewMod2"))
			Expect(*settings.Mods.WorkshopItems).To(Equal("111111;222222;333333;444444"))
		})

		It("should handle empty initial Mods strings", func() {
			settings := &zomboidv1.ZomboidSettings{
				WorkshopMods: []zomboidv1.WorkshopMod{
					{
						ModID:      ptr.To("NewMod1"),
						WorkshopID: ptr.To("333333"),
					},
				},
			}

			MergeWorkshopMods(settings)

			Expect(*settings.Mods.Mods).To(Equal("NewMod1"))
			Expect(*settings.Mods.WorkshopItems).To(Equal("333333"))
		})

		It("should handle empty WorkshopMods", func() {
			settings := &zomboidv1.ZomboidSettings{
				Mods: zomboidv1.Mods{
					Mods:          ptr.To("ExistingMod1;ExistingMod2"),
					WorkshopItems: ptr.To("111111;222222"),
				},
			}

			MergeWorkshopMods(settings)

			// Verify the existing mods remain unchanged
			Expect(*settings.Mods.Mods).To(Equal("ExistingMod1;ExistingMod2"))
			Expect(*settings.Mods.WorkshopItems).To(Equal("111111;222222"))
		})

		It("should handle nil ModID or WorkshopID", func() {
			settings := &zomboidv1.ZomboidSettings{
				WorkshopMods: []zomboidv1.WorkshopMod{
					{
						ModID:      ptr.To("NewMod1"),
						WorkshopID: nil,
					},
					{
						ModID:      nil,
						WorkshopID: ptr.To("444444"),
					},
				},
			}

			MergeWorkshopMods(settings)

			// Verify only non-nil values are merged
			Expect(*settings.Mods.Mods).To(Equal("NewMod1"))
			Expect(*settings.Mods.WorkshopItems).To(Equal("444444"))
		})
	})
})
