package words

import "math/rand"

var List = []string{
	"the", "of", "and", "a", "to", "in", "is", "you", "that", "it",
	"he", "was", "for", "on", "are", "as", "with", "his", "they", "at",
	"be", "this", "have", "from", "or", "one", "had", "by", "word", "but",
	"not", "what", "all", "were", "we", "when", "your", "can", "said", "there",
	"use", "each", "which", "she", "how", "their", "will", "way", "about", "many",
	"then", "them", "write", "would", "like", "so", "these", "her", "long", "make",
	"thing", "see", "him", "two", "has", "look", "more", "day", "could", "go",
	"come", "did", "number", "sound", "no", "most", "people", "over", "know", "water",
	"than", "call", "first", "who", "may", "down", "side", "been", "now", "find",
	"any", "new", "work", "part", "take", "get", "place", "made", "live", "where",
	"after", "back", "little", "only", "round", "man", "year", "came", "show", "every",
	"good", "give", "under", "name", "very", "through", "just", "form", "sentence", "great",
	"think", "say", "help", "low", "line", "differ", "turn", "cause", "much", "mean",
	"before", "move", "right", "boy", "old", "too", "same", "tell", "does", "set",
	"three", "want", "air", "well", "also", "play", "small", "end", "put", "home",
	"read", "hand", "port", "large", "spell", "add", "even", "land", "here", "must",
	"big", "high", "such", "follow", "act", "why", "ask", "men", "change", "went",
	"light", "kind", "off", "need", "house", "picture", "try", "again", "animal", "point",
	"mother", "world", "near", "build", "self", "earth", "father", "head", "stand", "own",
	"page", "should", "country", "found", "answer", "school", "grow", "study", "still", "learn",
	"plant", "cover", "food", "sun", "four", "between", "state", "keep", "eye", "never",
	"last", "let", "thought", "city", "tree", "cross", "farm", "hard", "start", "might",
	"story", "saw", "far", "sea", "draw", "left", "late", "run", "while", "press",
	"close", "night", "real", "life", "few", "north", "open", "seem", "together", "next",
	"white", "children", "begin", "got", "walk", "example", "ease", "paper", "group", "always",
	"music", "those", "both", "mark", "often", "letter", "until", "mile", "river", "car",
	"feet", "care", "second", "book", "carry", "took", "science", "eat", "room", "friend",
	"began", "idea", "fish", "mountain", "stop", "once", "base", "hear", "horse", "cut",
	"sure", "watch", "color", "face", "wood", "main", "enough", "plain", "girl", "usual",
	"young", "ready", "above", "ever", "red", "list", "though", "feel", "talk", "bird",
	"soon", "body", "dog", "family", "direct", "leave", "song", "measure", "door", "product",
	"black", "short", "numeral", "class", "wind", "question", "happen", "complete", "ship", "area",
	"half", "rock", "order", "fire", "south", "problem", "piece", "told", "knew", "pass",
	"since", "top", "whole", "king", "space", "heard", "best", "hour", "better", "true",
	"during", "hundred", "five", "remember", "step", "early", "hold", "west", "ground", "interest",
	"reach", "fast", "verb", "sing", "listen", "six", "table", "travel", "less", "morning",
	"ten", "simple", "several", "vowel", "toward", "war", "lay", "against", "pattern", "slow",
	"center", "love", "person", "money", "serve", "appear", "road", "map", "rain", "rule",
	"govern", "pull", "cold", "notice", "voice", "unit", "power", "town", "fine", "certain",
	"fly", "fall", "lead", "cry", "dark", "machine", "note", "wait", "plan", "figure",
	"star", "box", "noun", "field", "rest", "correct", "able", "pound", "done", "beauty",
	"drive", "stood", "contain", "front", "teach", "week", "final", "gave", "green", "oh",
	"quick", "develop", "ocean", "warm", "free", "minute", "strong", "special", "mind", "behind",
	"clear", "tail", "produce", "fact", "street", "inch", "multiply", "nothing", "course", "stay",
	"wheel", "full", "force", "blue", "object", "decide", "surface", "deep", "moon", "island",
	"foot", "system", "busy", "test", "record", "boat", "common", "gold", "possible", "plane",
	"stead", "dry", "wonder", "laugh", "thousand", "ago", "ran", "check", "game", "shape",
	"equate", "hot", "miss", "brought", "heat", "snow", "tire", "bring", "yes", "distant",
	"fill", "east", "paint", "language", "among", "grand", "ball", "yet", "wave", "drop",
	"heart", "am", "present", "heavy", "dance", "engine", "position", "arm", "wide", "sail",
	"material", "size", "vary", "settle", "speak", "weight", "general", "ice", "matter", "circle",
	"pair", "include", "divide", "syllable", "felt", "perhaps", "pick", "sudden", "count", "square",
	"reason", "length", "represent", "art", "subject", "region", "energy", "hunt", "probable", "bed",
	"brother", "egg", "ride", "cell", "believe", "fraction", "forest", "sit", "race", "window",
	"store", "summer", "train", "sleep", "prove", "lone", "leg", "exercise", "wall", "catch",
	"mount", "wish", "sky", "board", "joy", "winter", "sat", "written", "wild", "instrument",
	"kept", "glass", "grass", "cow", "job", "edge", "sign", "visit", "past", "soft",
	"fun", "bright", "gas", "weather", "month", "million", "bear", "finish", "happy", "hope",
	"flower", "clothe", "strange", "gone", "jump", "baby", "eight", "village", "meet", "root",
	"buy", "raise", "solve", "metal", "whether", "push", "seven", "paragraph", "third", "shall",
	"held", "hair", "describe", "cook", "floor", "either", "result", "burn", "hill", "safe",
	"cat", "century", "consider", "type", "law", "bit", "coast", "copy", "phrase", "silent",
	"tall", "sand", "soil", "roll", "temperature", "finger", "industry", "value", "fight", "lie",
	"beat", "excite", "natural", "view", "sense", "capital", "wont", "chair", "danger", "fruit",
	"rich", "thick", "soldier", "process", "operate", "guess", "necessary", "sharp", "wing", "create",
	"neighbor", "wash", "bat", "rather", "crowd", "corn", "compare", "poem", "string", "bell",
	"depend", "meat", "rub", "tube", "famous", "dollar", "stream", "fear", "sight", "thin",
}

func Generate(n int) []string {
	out := make([]string, n)
	prev := ""
	for i := 0; i < n; i++ {
		w := List[rand.Intn(len(List))]
		for w == prev {
			w = List[rand.Intn(len(List))]
		}
		out[i] = w
		prev = w
	}
	return out
}
