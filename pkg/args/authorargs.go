package args

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type AuthorArgs struct {
	Service  *string
	Protocol *string
	Cmd      *string
	CmdArg   []string `key:"cmd-arg"`
	ACL      *int
	InACL    *string
	OutACL   *string
	Addr     *string
	AddrPool *string
	Timeout  *int
	IdleTime *int
	AutoCmd  *string
	NoEscape *bool
	NoHangup *bool
	PrivLvl  *int `key:"priv-lvl"`
}

func (a *AuthorArgs) AsShellCommand() string {
    var command string
    if a.Cmd == nil {
        return ""
    }
    command = *a.Cmd
    for _, arg := range a.CmdArg {
        command = command + " " + arg
    }
    return command
}

func (a *AuthorArgs) GetService() string {
    if a.Service != nil {
        return *a.Service
    }
    return ""
}

// ParseAuthorArgs parses a slice of strings containing key=value pairs to an AuthorArgs struct
func ParseAuthorArgs(pairs []string) *AuthorArgs {
	args := &AuthorArgs{}

	for _, pair := range pairs {
		// The authorization arguments in both the REQUEST and the REPLY are
		// argument-value pairs. The argument and the value are in a single
		// string and are separated by either a "=" (0X3D) or a "*" (0X2A).
		// The equals sign indicates a mandatory argument. The asterisk
		// indicates an optional one.

        // First check for = sep
		parts := strings.SplitN(pair, "=", 2)
		var k string
		var v string
		if strings.ContainsRune(pair, '=') && len(parts) == 1 {
		    k = parts[0]
		    v = ""
	    } else if len(parts) == 2 {
    		k = parts[0]
    		v = parts[1]
		} else {
		    // Then check for * sep
		    parts = strings.SplitN(pair, "*", 2)
		    if strings.ContainsRune(pair, '*') && len(parts) == 1 {
		        k = parts[0]
		        v = ""
		    } else if len(parts) == 2 {
        		k = parts[0]
        		v = parts[1]
    		} else {
		        log.Printf("Could not parse argument: \"%v\"", pair)
		        continue // Skip to the next one
		    }
		}
// 		log.Printf("key: %v; value: %v\n", k, v)
		switch k {
		case "service":
			args.Service = &v
			break
		case "protocol":
			args.Protocol = &v
			break
		case "cmd":
			args.Cmd = &v
			break
		case "cmd-arg":
			args.CmdArg = append(args.CmdArg, v)
			break
		case "priv-lvl":
            log.Printf("Parsing priv-lvl: %v", v)
			lvl, err := strconv.ParseInt(v, 10, 0)
			if err != nil {
				log.Printf("Could not parse priv-lvl (%v): %v", v, err)
				continue
			}
			args.PrivLvl = Int(int(lvl))
			break
		case "noescape":
			args.NoEscape = BoolString(v)
			break
		case "nohangup":
			args.NoHangup = BoolString(v)
			break
		default:
			log.Printf("Unknown key: %v (ignoring)\n", v)
		}
	}

// 	fmt.Printf("Args: %+v\n", args)

	return args
}

func (args *AuthorArgs) Marshall() []string {
    sv := reflect.ValueOf(args).Elem()
	st := sv.Type()

	avpairs := make([]string, 0)

	for i := 0; i < sv.NumField(); i++ {
		svf := sv.Field(i)
		key := strings.ToLower(st.Field(i).Name)
		log.Printf("KEY ---- ", key)
		if alias, ok := st.Field(i).Tag.Lookup("key"); ok {
			if alias == "" {
			    log.Printf("Ignoring %v because the key tag is empty", key)
				continue // Empty key; ignore it?
			} else {
				key = alias
			}
		}
		log.Printf("KEY -++- ", key)

		if svf.Kind() == reflect.Ptr {
			if svf.IsNil() {
		        log.Printf("KEY %v is a nil Ptr", key)
		        if key == "priv-lvl" {
		            log.Printf("---- priv-lvl=%v", args.PrivLvl)
		        }
				continue
			}
			svf = svf.Elem()
		}

		val := ""

		switch svf.Kind() {
		case reflect.String:
			val = svf.String()
			break
		case reflect.Slice:
			for _, itm := range svf.Interface().([]string) {
				avpairs = append(avpairs, fmt.Sprintf("%v=%v", key, itm))
			}
			continue // already added avpairs!
		default:
			val = fmt.Sprintf("%v", svf.Interface())
			break
		}
		equalsOrAsterisk := "="
		if val == "" {
		    equalsOrAsterisk = "*"
		}
		avpairs = append(avpairs, fmt.Sprintf("%v%v%v", key, equalsOrAsterisk, val))
	}

	return avpairs
}
