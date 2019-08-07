package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/macroblock/imed/pkg/ptool"
)

var pdxRule = `
	entry                = '' scopeBody$;

	declr                = lval '=' rval [';'] [@comment];
	declrScope           = lval '=' scope [';'] [@comment];
	comparison           = lval @operators rval [';'] [@comment];
	list                 = @anyType {@anyType} [';']  [@comment];

	lval                 = @date|@int|@var|'"'#@string#'"';
	rval                 = @date|@hex|@percent|@var|@number|'"'#@string#'"';

	scope                = '{' (scopeBody|@empty) ('}'|empty);
	scopeBody            = (@declr|@declrScope|@comparison|@list){@declr|@declrScope|@comparison|@list};
	comment              = '#'#anyRune#{#!\x0a#!\x0d#!$#anyRune};

	int                  = ['-']digit#{#digit};
	float                = ['-'][int]#'.'#int;
	number               = float|int;
	percent              = int#'%'#'%';
	string               = {!'"'#stringChar};
	var                  = symbol#{#symbol};
	date                 = int#'.'#int#'.'#int#['.'#int];
	bool                 = 'yes'|'no';
	hex                  = '0x'#(digit|letter)(digit|letter)(digit|letter)(digit|letter)(digit|letter)(digit|letter)(digit|letter)(digit|letter);
	anyType              = number|percent|'"'#string#'"'|var|date|bool|hex;

	                     = {spaces|@comment};
	spaces               = \x00..\x20;
	anyRune              = \x00..$;
	digit                = '0'..'9';
	letter               = 'a'..'z'|'A'..'Z';
	operators            = '<'|'>';
	symbol               = digit|letter|'_'|':'|'@'|'.'|'-'|\u0027;
	stringChar           = ('\"'|anyRune);
	empty                = '';
`

var path = "c:/Users/admin/Documents/Paradox Interactive/Hearts of Iron IV/mod/oldworldblues_mexico/common/units/names"
var utf8bom = []byte{0xEF, 0xBB, 0xBF}
var pdx *ptool.TParser
var namesMap = make(map[string]Names)
var err error

type Names struct {
	Tag   string
	Ships map[string]*ShipType
}

type ShipType struct {
	Name    string
	Prefix  string
	Generic string
	Unique  []string
}

type NamesNew struct {
	Scope        string
	Name         string
	ForCountries []string
	ShipTypes    []string
	FallbackName string
	Unique       []string
}

type ShipName struct {
	Light      NamesNew
	Medium     NamesNew
	Heavy      NamesNew
	SuperHeavy NamesNew
}

func main() {
	// Build parser.
	pdx, err = ptool.NewBuilder().FromString(pdxRule).Entries("entry").Build()
	if err != nil {
		panic(err)
	}

	files, err := filepath.Glob(filepath.FromSlash(path) + string(os.PathSeparator) + "*.txt")
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		if strings.HasSuffix(f, "00_names.txt") {
			continue
		}

		err = parseNamesFile(f)
		if err != nil {
			panic(err)
		}
	}

	for _, n := range namesMap {
		var sn ShipName
		for _, t := range n.Ships {
			var r NamesNew
			r.ForCountries = append(r.ForCountries, n.Tag)
			r.FallbackName = t.Generic + " %d"
			r.Unique = append(r.Unique, t.Unique...)
			sort.Strings(r.Unique)
			switch t.Name {
			case "destroyer_equipment":
				r.Scope = strings.ToUpper(n.Tag) + "_LIGHT"
				r.Name = "NAME_THEME_LIGHT"
				r.ShipTypes = append(r.ShipTypes, "light_ship_unit")
				sn.Light = r
			case "light_cruiser_equipment":
				r.Scope = strings.ToUpper(n.Tag) + "_MEDIUM"
				r.Name = "NAME_THEME_MEDIUM"
				r.ShipTypes = append(r.ShipTypes, "medium_ship_screen_unit")
				sn.Medium = r
			case "heavy_cruiser_equipment":
				r.Scope = strings.ToUpper(n.Tag) + "_HEAVY"
				r.Name = "NAME_THEME_HEAVY"
				r.ShipTypes = append(r.ShipTypes, "heavy_ship_unit")
				sn.Heavy = r
			case "battleship_equipment":
				r.Scope = strings.ToUpper(n.Tag) + "_SUPER_HEAVY"
				r.Name = "NAME_THEME_SUPER_HEAVY"
				r.ShipTypes = append(r.ShipTypes, "super_heavy_ship_unit")
				sn.SuperHeavy = r
			case "carrier_equipment":
				// skip
			}
		}

		output := string(utf8bom)
		r := sn.Light
		output += r.Scope + " = {\r\n\tname = " + r.Name + "\r\n\tfor_countries = { " + r.ForCountries[0] + " }\r\n\ttype = ship\r\n\tship_types = { " + r.ShipTypes[0] + " }\r\n\tfallback_name = \"" + r.FallbackName + "\"\r\n\tunique = {\r\n\t\t" + strings.Join(r.Unique, " ") + "\r\n\t}\r\n}\r\n\r\n"
		r = sn.Medium
		output += r.Scope + " = {\r\n\tname = " + r.Name + "\r\n\tfor_countries = { " + r.ForCountries[0] + " }\r\n\ttype = ship\r\n\tship_types = { " + r.ShipTypes[0] + " }\r\n\tfallback_name = \"" + r.FallbackName + "\"\r\n\tunique = {\r\n\t\t" + strings.Join(r.Unique, " ") + "\r\n\t}\r\n}\r\n\r\n"
		r = sn.Heavy
		output += r.Scope + " = {\r\n\tname = " + r.Name + "\r\n\tfor_countries = { " + r.ForCountries[0] + " }\r\n\ttype = ship\r\n\tship_types = { " + r.ShipTypes[0] + " }\r\n\tfallback_name = \"" + r.FallbackName + "\"\r\n\tunique = {\r\n\t\t" + strings.Join(r.Unique, " ") + "\r\n\t}\r\n}\r\n\r\n"
		r = sn.SuperHeavy
		output += r.Scope + " = {\r\n\tname = " + r.Name + "\r\n\tfor_countries = { " + r.ForCountries[0] + " }\r\n\ttype = ship\r\n\tship_types = { " + r.ShipTypes[0] + " }\r\n\tfallback_name = \"" + r.FallbackName + "\"\r\n\tunique = {\r\n\t\t" + strings.Join(r.Unique, " ") + "\r\n\t}\r\n}\r\n"

		err = ioutil.WriteFile(strings.ToUpper(n.Tag)+"_ship_names.txt", []byte(output), 0755)
		if err != nil {
			panic(err)
		}
	}
}

func trimQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func readFile(path string) (string, error) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	// Remove utf-8 bom if found.
	if bytes.HasPrefix(f, utf8bom) {
		f = bytes.TrimPrefix([]byte(f), utf8bom)
	}
	return string(f), nil
}

func parseNamesFile(path string) error {
	fmt.Println(path)
	f, err := readFile(path)
	if err != nil {
		return err
	}
	// fmt.Println(f)
	if len(f) > 0 {
		node, err := pdx.Parse(f)
		if err != nil {
			return err
		}
		_ = node
		// fmt.Println(ptool.TreeToString(node, pdx.ByID))
		var n Names
		n.Tag = node.Links[0].Links[0].Value
		n.Ships = make(map[string]*ShipType)
		namesMap[n.Tag] = n
		// fmt.Println(node.Links[0].Links[0].Value)
		err = traverseNamesFile(node, n)
		if err != nil {
			return err
		}

	}
	return nil
}

func traverseNamesFile(root *ptool.TNode, n Names) error {
	for _, node := range root.Links {
		nodeType := pdx.ByID(node.Type)
		switch nodeType {
		case "declrScope":
			switch strings.ToLower(node.Links[0].Value) {
			case "destroyer_equipment", "light_cruiser_equipment", "heavy_cruiser_equipment", "battleship_equipment", "carrier_equipment":
				n.Ships[strings.ToLower(node.Links[0].Value)] = &ShipType{Name: strings.ToLower(node.Links[0].Value)}
				for _, link := range node.Links {
					nodeType := pdx.ByID(link.Type)
					switch nodeType {
					case "declr":
						switch strings.ToLower(link.Links[0].Value) {
						case "prefix":
							n.Ships[strings.ToLower(node.Links[0].Value)].Prefix = trimQuotes(link.Links[1].Value)
						}
					case "declrScope":
						switch strings.ToLower(link.Links[0].Value) {
						case "generic":
							for _, link := range link.Links {
								nodeType := pdx.ByID(link.Type)
								switch nodeType {
								case "list":
									for _, link := range link.Links {
										n.Ships[strings.ToLower(node.Links[0].Value)].Generic = trimQuotes(link.Value)
									}
								}
							}
						case "unique":
							for _, link := range link.Links {
								nodeType := pdx.ByID(link.Type)
								switch nodeType {
								case "list":
									for _, link := range link.Links {
										n.Ships[strings.ToLower(node.Links[0].Value)].Unique = append(n.Ships[strings.ToLower(node.Links[0].Value)].Unique, link.Value)
									}
								}
							}
						}
					}
				}
			default:
				err := traverseNamesFile(node, n)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
