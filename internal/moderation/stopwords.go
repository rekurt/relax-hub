package moderation

// Stopwords — list of prohibited words with variations
var Stopwords = []WordVariation{
	// Common profanity (Russian and transliterations)
	{
		base:       "пиздец",
		variations: []string{"пиздец", "пиздец", "pezdec", "pizdes", "п1здец", "п!здец"},
	},
	{
		base:       "пизда",
		variations: []string{"пизда", "пидза", "pizda", "pidza", "п1зда", "п!зда"},
	},
	{
		base:       "хуй",
		variations: []string{"хуй", "hui", "х0й", "х@й", "huй"},
	},
	{
		base:       "блядь",
		variations: []string{"блядь", "blyad", "б1ядь", "б!ядь"},
	},
	{
		base:       "ебать",
		variations: []string{"ебать", "ebat", "е6ать", "е@ать"},
	},
	{
		base:       "сука",
		variations: []string{"сука", "suka", "с0ка", "су6а"},
	},
	{
		base:       "гавно",
		variations: []string{"гавно", "gavno", "г@вно", "га6но", "г4вн0", "гавн0"},
	},
	{
		base:       "срань",
		variations: []string{"срань", "sran", "ср@нь", "сра6ь"},
	},
	{
		base:       "дрочить",
		variations: []string{"дрочить", "drocit", "др0чить"},
	},
	{
		base:       "ебаный",
		variations: []string{"ебаный", "ebaniy", "е6аный"},
	},
	{
		base:       "хуева",
		variations: []string{"хуева", "hueva", "х0ева"},
	},
	{
		base:       "ебаться",
		variations: []string{"ебаться", "ebatsya", "е6аться"},
	},
}

type WordVariation struct {
	base       string
	variations []string
}

func (w WordVariation) Variations() []string {
	return w.variations
}

func (w WordVariation) Base() string {
	return w.base
}
