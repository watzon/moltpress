# Profile Theming Learnings

## Migration Pattern
- Migrations are defined in `internal/database/migrations.go` as a slice of structs
- Each migration has a `name` and `sql` field
- Use `IF NOT EXISTS` for idempotent SQL
- Format: `ALTER TABLE users ADD COLUMN IF NOT EXISTS column_name TYPE DEFAULT value`

## 004_add_theme_settings Migration
- Added `theme_settings JSONB DEFAULT NULL` column to users table
- NULL means "use defaults" - no default JSON object needed
- No NOT NULL constraint to allow null values
- Build verified: `go build ./...` passes
## 2026-02-01: ThemeSettings Implementation

### Patterns Used
- Pointer types for nullable fields (`*string`, `*bool`) following codebase convention
- JSON tags with `omitempty` for optional fields
- Sentinel errors at package level for validation failures
- Case-insensitive font preset matching using `strings.EqualFold()`

### CSS Sanitization Approach
- Size limit: 10KB max to prevent abuse
- Blocked patterns checked first (regex for dangerous CSS)
- Property allowlist: 22 safe CSS properties only
- Rules parsed by semicolon delimiter
- Property names normalized to lowercase for matching
- Returns sanitized CSS with only allowed properties

### Validation Functions
- `IsValidFontPreset()`: Checks against 10 preset font names
- `IsValidHexColor()`: Validates #RGB or #RRGGBB format
- `ThemeSettings.Validate()`: Comprehensive validation of all theme fields

### Security Considerations
- Blocked: url(), @import, expression(), javascript:, -moz-binding, behavior:, position: fixed/absolute
- All user CSS is sanitized before storage
- Invalid/malformed CSS rules are silently dropped
## 2026-02-01: Repository & API Integration

### JSONB Handling Pattern
- PostgreSQL JSONB columns scanned into `[]byte` variable
- Unmarshal to struct after successful scan: `json.Unmarshal(themeJSON, user.ThemeSettings)`
- Marshal before insert/update: `json.Marshal(merged)`
- Use `COALESCE($n, column_name)` for partial updates preserving existing

### Deep Merge Strategy
- `MergeThemeSettings(existing, update)` preserves existing values
- Update values take precedence (overwrite existing if provided)
- Each sub-struct (Colors, Fonts, Toggles) merged individually
- CustomCSS overwrites completely (no merge semantics)
- nil existing + nil update = nil result

### Validation in Handler vs Repository
- Handler validates early for 400 Bad Request with user-friendly message
- Repository also validates as safety net (defense in depth)
- Handler catches theme validation errors from repo and returns 400
- Unhandled errors return 500 Internal Server Error

### API Update Flow
1. Handler receives PATCH /api/v1/me with theme_settings
2. Handler validates with `req.ThemeSettings.Validate()`
3. Repository fetches existing theme_settings from DB
4. Repository sanitizes CustomCSS if provided
5. Repository merges existing + update
6. Repository stores merged JSON, returns user with theme
7. Handler returns `user.ToPublic()` with theme_settings

### Key Files Modified
- `internal/users/model.go`: Added ThemeSettings to User, UserPublic, UpdateUserRequest, ToPublic()
- `internal/users/theme.go`: Added MergeThemeSettings()
- `internal/users/repository.go`: Update(), GetByUsername(), GetWithStats() handle JSONB
- `internal/api/handlers.go`: handleUpdateMe validates theme, handles validation errors

## 2026-02-01: ThemeSettings TypeScript Interface

### Changes Made
- Added `ThemeColors`, `ThemeFonts`, `ThemeToggles`, and `ThemeSettings` interfaces to `web/src/lib/api/client.ts`
- Extended `User` interface with optional `theme_settings?: ThemeSettings` field
- Updated `updateMe` method to accept `theme_settings` in its parameter type

### Interface Structure
```typescript
export interface ThemeColors {
  background?: string;
  text?: string;
  accent?: string;
  link?: string;
  title?: string;
}

export interface ThemeFonts {
  title?: string;
  body?: string;
}

export interface ThemeToggles {
  show_avatar?: boolean;
  show_stats?: boolean;
  show_follower_count?: boolean;
  show_bio?: boolean;
}

export interface ThemeSettings {
  colors?: ThemeColors;
  fonts?: ThemeFonts;
  toggles?: ThemeToggles;
  custom_css?: string;
}
```

### Patterns Observed
- All optional fields use `?:` syntax consistently
- Interfaces are exported for use in components
- The API client uses PATCH `/me` endpoint for profile updates
- Type checking passes with `npm run check` (0 errors)

### Verification
- `cd web && npm run check` passes with 0 errors
- 6 pre-existing accessibility warnings remain (unrelated to this change)


## 2026-02-01: SKILL.md Documentation

### Documentation Added
Added comprehensive "## Profile Theming" section to SKILL.md at line 128 (after "User Profiles" section).

### Section Structure
1. **Setting Theme Colors** - Hex color customization for background, text, accent, link, title
2. **Font Presets** - Table of 10 curated font options with descriptions
3. **Custom CSS** - Allowed properties list and security-blocked patterns
4. **Toggle Options** - Show/hide profile elements (avatar, stats, follower count, bio)
5. **Complete Theme Example** - Full JSON showing all options together
6. **Partial Updates** - How PATCH merges with existing settings
7. **Resetting Theme** - Setting theme_settings to null restores defaults

### Key Patterns Used
- Consistent curl format matching existing SKILL.md style
- All examples use `$MOLTPRESS_API_KEY` environment variable
- JSON payloads formatted for readability with proper indentation
- Security considerations clearly documented for CSS
- Practical examples showing real-world use cases

### Font Presets Documented (10 total)
- inter (default), georgia, playfair, roboto, lora
- montserrat, merriweather, source-code-pro, oswald, raleway

### CSS Security Model
- Allowlist: 25 safe CSS properties
- Blocklist: Dangerous patterns (url(), @import, expression(), etc.)

### Verification
- Section added at line 128 in SKILL.md
- `grep "Profile Theming" SKILL.md` returns content successfully
- All subsections render correctly in markdown
