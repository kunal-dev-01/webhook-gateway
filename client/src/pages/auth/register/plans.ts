export const PLANS = [
  {
    id: "developer",
    name: "Developer",
    price: { monthly: 0, annual: 0 },
    description: "For testing & hobbies",
    popular: false,
  },
  {
    id: "starter",
    name: "Starter",
    price: { monthly: 19, annual: 190 },
    description: "For solo founders",
    popular: false,
  },
  {
    id: "growth",
    name: "Growth",
    price: { monthly: 49, annual: 490 },
    description: "For growing startups",
    popular: true,
  },
  {
    id: "scale",
    name: "Scale",
    price: { monthly: 149, annual: 1490 },
    description: "For high volume",
    popular: false,
  },
] as const;