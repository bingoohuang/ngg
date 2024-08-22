# ss

1. `func If[T any](condition bool, a, b T) T`
2. `func IfFunc[T any](condition bool, a, b func() T) T `
3. `func ToSet[K comparable](v []K) map[K]bool`
4. `func Parse[T Parseable](str string) (T, error)`
5. `func IndexN(s, sep string, n int) int`
6. `func Must[A any](a A, err error) A`
7. `func Or[T comparable](a, b T) T`
8. `func FnMatch(pattern, name string, caseInsensitive bool) (matched bool, err error)` 文件名模式匹配
9. `Contains`, `AnyOf`, `IndexOf`, `Split`, `Split2`, `HasPrefix`, `HasSuffix`
10. `Abbreviate` 缩略
11. `QuoteSingle`, `UnquoteSingle` 单引号引用
12. `ss.Base64().Encode/Decode` base64 编码/解码

## 中国身份证等随机信息

```go
fmt.Println("姓名:", ss.Rand().ChineseName())
fmt.Println("性别:", ss.Rand().Sex())
fmt.Println("地址:", ss.Rand().Address())
fmt.Println("手机:", ss.Rand().Mobile())
fmt.Println("身份证:", ss.Rand().ChinaID())
fmt.Println("有效期:", ss.Rand().ValidPeriod())
fmt.Println("发证机关:", ss.Rand().IssueOrg())
fmt.Println("邮箱:", ss.Rand().Email())
fmt.Println("银行卡:", ss.Rand().BankNo())
fmt.Println("日期:", ss.Rand().Time())
```

```
姓名: 武锴脹
性别: 男
地址: 四川省攀枝花市嫯航路3755号婘螐小区3单元1216室
手机: 18507708621
身份证: 156315197605103397
有效期: 20020716-20220716
发证机关: 平凉市公安局某某分局
邮箱: wvcykkyh@kjsth.co
银行卡: 6230959897028597497
日期: 1977-06-16 23:41:28 +0800 CST
```

### Resources

1. [Chinese Id Card Number (Resident Identity Card) and name Generator](https://www.myfakeinfo.com/nationalidno/get-china-citizenidandname.php)
2. [China ID](https://github.com/mritd/chinaid)


## strcase

forked from https://github.com/iancoleman/strcase

strcase is a go package for converting string to various cases (e.g. [snake case](https://en.wikipedia.org/wiki/Snake_case) or [camel case](https://en.wikipedia.org/wiki/CamelCase)) to see the full conversion table below.

### Example


| s                      | Function                   | Result                   |
|------------------------|----------------------------|--------------------------|
| `AnyKind of string v5` | `ToSnake(s)`               | `any_kind_of_string_v5`  |
| `AnyKind of string v5` | `ToSnakeUpper(s)`          | `ANY_KIND_OF_STRING_V5`  |
| `AnyKind of string v5` | `ToKebab(s)`               | `any-kind-of-string-v5`  |
| `AnyKind of string v5` | `ToKebabUpper(s)`          | `ANY-KIND-OF-STRING5-V5` |
| `AnyKind of string v5` | `ToDelimited(s, '.')`      | `any.kind.of.string.v5`  |
| `AnyKind of string v5` | `ToDelimitedUpper(s, '.')` | `ANY.KIND.OF.STRING.V5`  |
| `AnyKind of string v5` | `ToCamel(s)`               | `AnyKindOfStringV5`      |
| `mySQL`                | `ToCamel(s)`               | `MySql`                  |
| `AnyKind of string v5` | `ToCamelLower(s)`          | `anyKindOfStringV5`      |
| `ID`                   | `ToCamelLower(s)`          | `id`                     |
| `SQLMap`               | `ToCamelLower(s)`          | `sqlMap`                 |
| `TestCase`             | `ToCamelLower(s)`          | `fooBar`                 |
| `foo-bar`              | `ToCamelLower(s)`          | `fooBar`                 |
| `foo_bar`              | `ToCamelLower(s)`          | `fooBar`                 |


case conversion types:

- Camel Case (e.g. CamelCase)
- Lower Camel Case (e.g. lowerCamelCase)
- Snake Case (e.g. snake_case)
- Screaming Snake Case (e.g. SCREAMING_SNAKE_CASE)
- Kebab Case (e.g. kebab-case)
- Screaming Kebab Case(e.g. SCREAMING-KEBAB-CASE)
- Dot Notation Case (e.g. dot.notation.case)
- Screaming Dot Notation Case (e.g. DOT.NOTATION.CASE)
- Title Case (e.g. Title Case)
- Other delimiters

### resources

1. [caps a case conversion library for Go](https://github.com/chanced/caps)
