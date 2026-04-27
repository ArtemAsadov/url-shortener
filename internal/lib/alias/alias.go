package alias

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

// Generator отвечает за генерацию алиасов с настройками
type Generator struct {
	MaxLength int
}

// New создаёт новый генератор с заданной максимальной длиной алиаса
func New(maxLength int) *Generator {
	return &Generator{MaxLength: maxLength}
}

// GenerateFromURL — основной метод: создаёт алиас из URL, используя настройки генератора
func (g *Generator) GenerateFromURL(urlStr string) string {
	// Используем g.MaxLength вместо внешнего параметра
	aliasLength := g.MaxLength
	if aliasLength <= 0 {
		aliasLength = 10 // дефолт, если не настроен
	}

	// 1. Парсим через regex
	groups := regexpNameGroups(urlPattern, urlStr)
	if groups == nil {
		return g.generateHash(urlStr, aliasLength)
	}

	// 2. Берём домен
	clean := groups["domain"]

	// 3. Санитизация
	clean = cleanRe.ReplaceAllLiteralString(strings.ToLower(clean), "-")
	clean = strings.Trim(clean, "-")
	clean = dashRe.ReplaceAllLiteralString(clean, "-")

	// 4. Фоллбэк если пусто
	if clean == "" {
		return g.generateHash(urlStr, aliasLength)
	}

	// 5. Обрезка по лимиту
	// ⚠️ Внимание: len() считает байты, а не руны (см. примечание ниже)
	if len(clean) > aliasLength {
		clean = clean[:aliasLength]
		clean = strings.TrimRight(clean, "-")
	}

	return clean
}

// generateHash — вспомогательный метод (приватный, маленькая буква)
func (g *Generator) generateHash(str string, length int) string {
	hash := sha256.Sum256([]byte(str))
	hexStr := hex.EncodeToString(hash[:])
	if len(hexStr) > length {
		return hexStr[:length]
	}
	return hexStr
}

var urlPattern = regexp.MustCompile(`^(?P<scheme>https?)://(?P<domain>[^/]+)(?P<path>/.*)?$`)

var (
	cleanRe = regexp.MustCompile(`[^a-z0-9.-]+`)
	dashRe  = regexp.MustCompile(`-+`)
)

// regexpNameGroups — утилитарная функция (может остаться глобальной, т.к. не зависит от состояния)
func regexpNameGroups(re *regexp.Regexp, s string) map[string]string {
	match := re.FindStringSubmatch(s)
	if match == nil {
		return nil
	}

	names := re.SubexpNames()
	result := make(map[string]string, len(names)-1)

	for i, name := range names {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result
}
