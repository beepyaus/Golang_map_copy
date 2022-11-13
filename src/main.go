package main

import (
"fmt"
"io/fs"
"io"
"os"
"os/exec"
"flag" 
"log"
"bytes" 
"strings"
"errors"
"regexp"
"encoding/xml"
"encoding/csv"
"strconv"
"time" 
)

// NOTE: 'st' terminal (Suckless) did NOT pickup the proper formatting of comments for some reason. 
//        Alacritty seems fine as of Sun 13 Nov 2022 13:00:24
// #############################################################################################################
// # part 1. 
// #   - should of copied (unix cp) the base and specific base_variant dir into the /tmp swap dir
// #   - re-created the failsafe default http website pages 
// #   - NO chown should of be done yet as Vagrant ( or Docker ) 
// #     does not have permission to chown the 'troy' user owned files 
// #
// # part 2 - this file (ruby/perl/rust version
// #   - chown and chmod the swap/temp dir to correct ACLs etc 
// #   - use the XML spec file for lookup , normally
// #     on dev machine: ~/Development/Jobi/Utils/sync/assets/config/base_TREE_SPECS/spec_foo.xml
// #     on live machine: ~/sync/assets/config/base_TREE_SPECS/spec_foo.xml
//       
// #   - refer to the file-system file area normally 
// #     on dev machine:  ~/Development/Jobi/Utils/sync/assets/config/base_{PROD,OTHER_TAG}
// #     on live machine:  ~/sync/assets/config/base_{PROD,OTHER_TAG}
// #   - cross-reference the spec xml file with the file-system space. 
// #   - ANY directory not listed in the XML spec file cannot be copied over to the target machine. 
// #   - rsync the files across. 
// #   - simple_copy: is for straight forward folders, where the XML spec is too much work 
// #   - map_copy: uses the XML spec and it the main point of this project. 
// 
// #   - NOTE: to be run on PROD/AWS EC2 or Vagrant and not really for the Dev Machine!!
// #     Hence the TEST_PREFIX const! 
// 
// #   - perform rsync or similar on all the required dirs/files into the target (normally LIVE/PROD) server!!!
// 
// #   TODO: 
// #        - handle rsync stdout response 
// #        - OO ? wrap in a class? 
// ###############################################################################################################

var Version string = "0.0.1"

// #Hardcoded value to prefix the target destination for testing
// #SET TO "" for the LIVE/REAL scenario testing
var TEST_PREFIX = "/home/troy/Downloads/golang_test_mapcopy"
//TEST_PREFIX = ""

var Configdir string
var Swapdir string 
var Target string 
var Buildname string 
var Builddir string 
var Sourcedir string 
var Backupdir string 
var Logfiledir string 
var DryRun bool 
var ForceYes bool 
var RunMode string
var BypassTargetNull bool 
var Debug int //really log_level

type Command struct {
//c_ only cause some are reserved words .
    c_type string 
    c_path string 
    c_delete bool 
    c_user string 
    c_group string 
}
type FileData struct {
    node string //'name' attr in XML spec. added cause of firehose model
    file_level int
    file_type string
    file_user string
    file_group string
    file_mode string
    default_file_user string
    default_file_group string
    default_file_mode string
}

//recursive added map of each file and dir
//xml treespec version
var FileMap map[string]FileData

//recursive added map of each file and dir 
//filesystem version
var FileSourceMap map[string]FileData

func main() {
    s := "good bye"
    var p *string = &s
    *p = "ciao"
    fmt.Printf("Here is the pointer p: %p\n", p) // prints address
    fmt.Printf("Here is the string *p: %s\n", *p) // prints string
    fmt.Printf("Here is the string s: %s\n", s) // prints same string

    FileMap = make(map[string]FileData )
    FileSourceMap = make(map[string]FileData )

    var terminate bool 

    terminate, DryRun, ForceYes, RunMode, BypassTargetNull, Debug = set_args() 

    //fmt.Printf("terminate %v \n" , terminate)
    if terminate == true {
        return 
    }

    //fmt.Printf("Debug = %v:\n" , Debug ) 

    Debugln("testing one two three")
    Errorln("testing one two three")
    Infoln("testing one two three")

    Infoln(fmt.Sprintf("terminate: %v", terminate))
    Infoln(fmt.Sprintf("dry_run: %v", DryRun), "underline") 
    Infoln(fmt.Sprintf("force_yes: %v", ForceYes))
    Infoln(fmt.Sprintf("run_mode: %v", RunMode))
    Infoln(fmt.Sprintf("bypass_target_null: %v", BypassTargetNull))
    Infoln(fmt.Sprintf("debug: %v", Debug))

    res_globals, err := set_globals()
    if err != nil {
        Errorln( fmt.Sprintf("set_globals failed!: %v", err.Error() ), "bold") 
        Errorln( fmt.Sprintf("res_globals: %v", res_globals )) ; 
        return 
    }

    splash(Version) 

    //TODO
    // if res , err := clean_backup_dir(Backupdir); err != nil {
    //    Errorln("clean_backup_dir failure!") 
    //    Errorln("clean_backup_dir: " + res ) 
    //    return 
    // }

    // if res , err := setup_logfile_dir(Logfiledir); err != nil {
    //    Errorln("setup_logfile_dir failure!") 
    //    Errorln("setup_logfile_dir: " + err.Error() ) 
    //    Infoln( fmt.Sprintf("setup_logfile_dir result: %v ", res))
    //    return 
    // }
    
    webowner := get_webowner()
    Debugln("webowner: " + webowner) 

    map_copy("/root", false)

// 
// command_list, err = create_command_list($configdir, mapcopy_csv_file) 
// unless err == nil 
//   error "create_command_list failed!" 
//   exit 5
// end 
// 
// #TODO Need to decide if another ARGV is needed to proceed-on-error 
// # or terminate of the first/any error and ignore the outstanding items in list.
// 
// unless true == result = process_commands(command_list)
//   error "process_commands failed!"
//   exit 6
// end 

}

// type my_error struct {
//     my_error_str string 
// }
// 
// func (e *my_error) Error() string {
//     return "my_error:: " + e.my_error_str
// }

func run_command(cmd *exec.Cmd) ( string, error)  {

    var out bytes.Buffer
    var errout bytes.Buffer

    //TODO: an err may NOT happen and 
    // the stderr output probabaly should be processed etc better
    cmd.Stderr = &errout
    cmd.Stdout = &out 

    err := cmd.Run()
	if err != nil {
		//log.Fatal(err)
        log.Printf("Command finished with error: %v", err)
        Errorln( fmt.Sprintf("Cmd finished with error: %v", err) , "bold")
        Errorln( "Stderr: " + errout.String()  )

        myErr := errors.New(err.Error()) 
    
        return "ERROR", myErr
	}

    return out.String(), nil
}

func get_platform() string { 
//remember the -s switch to just get the OS name    

    cmd := exec.Command("uname", "-a") 

    result, err := run_command(cmd ) 
    
    if err != nil {
        Errorln( fmt.Sprintf( "Error calling run command : %v " , err) )
        return "real error"
    }

    return result 

}

func get_log_level(level string) int {
    switch level {
        case "error":
            return 1

        case "debug":
            return 2

        case "info" :
            return 3 

        default:
            return 0 
    }
} 

func set_args() ( bool, bool, bool, string, bool, int ) {
// when terminate < 0 the handler NEEDS to kill this process!!!!!
// - shift will pop first element and move the whole array to the left. 
//NOTE: MAYBE cause it is  "go run .." but the vars did NOT seem to be set correctly! 
// ..I refreshed a few times and then it worked. but I swear the code was right. 
    
    var help bool
    var version bool

    //return set 
    var terminate bool 
    var dry_run bool
    var force_yes bool
    var run_mode string
    var bypass_target_null bool
    var log_level int
    var log_level_text string
    
    flag.BoolVar(&dry_run, "dry-run", false, "either 'echo ' prefix for '--dry-run' flag for rsync,")
    flag.BoolVar(&dry_run, "d", false, "either 'echo ' prefix for '--dry-run' flag for rsync,")
    flag.BoolVar(&force_yes, "force-yes", false, "auto yes when stdin asking for user input.")
    flag.BoolVar(&force_yes, "f", false, "auto yes when stdin asking for user input.")
    flag.BoolVar(&help, "help", false, "")
    flag.BoolVar(&help, "h", false, "")
    flag.StringVar(&run_mode, "mode", "", "")
    flag.StringVar(&run_mode, "m", "", "")
    flag.BoolVar(&version, "version", false, "")
    flag.BoolVar(&version, "v", false, "")
    flag.StringVar(&log_level_text , "l", "", "")
    flag.StringVar(&log_level_text, "log-level","" , "")
    flag.BoolVar(&bypass_target_null, "b", false, "")
    flag.BoolVar(&bypass_target_null, "bypass-null", false, "")
    flag.Parse()

    if help || version { 
        terminate = true
    }
    if help { 
        show_help() 
    }
    if version {
        show_version()
    }

    //fmt.Printf("get_log_level(log_level) = %v " , get_log_level(log_level_text))

    log_level = get_log_level(log_level_text)

    return terminate, dry_run, force_yes, run_mode, bypass_target_null, log_level

//-------------------------------OR----------------------------------
//this does work , but the 'flag' package SEEMS to work. 
//commented out as Go doesnt like vars no used ...like Rust I suppose
    //args := os.Args

    // for ; len(args) > 0 ;  {
    //     
    //     arg := args[0] 
    //     //doesnt kick in until log_level set anyways...
    //     //debug( "set_args: ind0: '#{arg}' " )
    //     switch arg {
    //         case "--dry-run", "-d":
    //             dry_run = true
    //         case "--force-yes","-f":
    //           force_yes = true
    //         case "--mode","-m":
    //           run_mode = args[1] 
    //           case "--help","-h" : 
    //           show_help() 
    //           terminate = true
    //           case "--version", "-v" : 
    //           show_version()
    //           terminate = true
    //           case "--log-level", "-l" : 
    //           debug = get_log_level(args[1]) 
    //           case "--bypass-null", "-b" : 
    //           bypass_target_null = true
    //       } 
    //     args = args[1:]
    // }
    // -------------------------------------------------------------------

}

func get_debug() int { 
   return Debug
}


func parse_debug_opts(opts []string) (bool,bool){ 

    var bold_me = false 
    var underline = false 

    for _, v := range opts {
        switch v {
            case "bold": 
                bold_me = true
            case "underline": 
                underline = true
        }
    }

    return bold_me, underline 
}

func Say(data string, opts ...string) { 
//Perl say 

    bold, underline := parse_debug_opts(opts) 
    color := ""
    prefix := ""

    var style string 
    if bold == true { 
         style = "1" 
    } else if underline {
        style = "4"
    }else{
        style = ""
    }

    fmt.Printf("\033[%s;%sm%s: %s  \033[0m\n", style,  color, prefix,  data)
}

func Debugln(data string, opts ...string) { 

    bold, underline := parse_debug_opts(opts) 
    yellow := "33"
    debug_writer(data, yellow, bold, "DEBUG:", 2, underline) 
}

func Errorln(data string, opts ...string) { 

    bold, underline := parse_debug_opts(opts) 
    red := "31"
    debug_writer(data, red, bold, "ERROR:", 1, underline)
}

func Infoln(data string, opts ...string) { 

    bold, underline := parse_debug_opts(opts) 
    blue := "34"
    debug_writer(data, blue, bold, "INFO:", 3, underline)
}

func debug_writer(data string, color string, bold_me bool, prefix string, writer_level int, underline_me bool ) {

    var style string 
    if bold_me == true { 
         style = "1" 
    }else if  underline_me {
        style = "4"
    }else{
        style = ""
    }

    if get_debug() >= writer_level {
        fmt.Printf("\033[%s;%sm%s: %s  \033[0m\n", style,  color, prefix,  data)
    }
}

func splash(version string) { 

    fmt.Printf(`

        &&& &&  & &&
       && &\/&\|& ()|/ @, &&
       &\/(/&/&||/& /_/)_&/_&
    &() &\/&|()|/&\/ '%%" & ()
   &_\_&&_\ |& |&&/&__%%_/_& &&
 &&   && & &| &| /& & %% ()& /&&
  ()&_---()&\&\|&&-&&--%%---()~
      &&     \|||
              |||               mapcopy XML Tree Spec File copier. 
              |||               Go Version 1.x
              |||               Version: %s
        , -=-~  .-^- _ 
 
        %s`, version, "\n")

}



func show_version() {
    fmt.Println("Version " + Version) 
}

func show_help() {
    fmt.Printf(`
SYNOPSIS
    mapcopy.go OPTIONS 

DESCRIPTION
    copies via rsync calls all of the listed /etc, config, other files 
    with their proper file mode (mode/user/group) to the target dirs referenced 
    in the XML Tree Specification files. 
    
    OPTIONS
        -h, --help 
        show this help area. 

        -d, --dry-run 
        prefix the critical rm -rf, rsync calls with either "echo " or use their own dry-run flag 
        to avoid doing a real operation that will delete or move or change files etc. 

        -m, --mode {LIVE|DEV|TEST} ...or other 
        tell the script what to do in certain events/operations. 
        currently not actually in use -only originally the dry-run option was flagged when NOT 'live' 

        -f, --force-yes 
        when stdin is asking for a y/n question force a 'y' to continue without user interaction

        -l, --log-level {debug|error|other}
        output extra debug 'puts' lines when needed ...like Rust's logging. 

        -b, --bypass-null 
        allow a target prefix of 'NULL', normally meaning it was the Development machine 
        Warning: allowed to run on a live production machine, would copy over Development settings 
        to the live working directories!

        Last Edited: Tue 08 Nov 2022 21:18:53 

        %s`, "\n" )

}

// #Terminal colors 
// #Foreground Code	Background Code
// #  Black
// #  	30	40
// #  Red
// #  	31	41
// #  Green
// #  	32	42
// #  Yellow
// #  	33	43
// #  Blue
// #  	34	44
// #  Magenta
// #  	35	45
// #  Cyan
// #  	36	46
// #  White
// #  	37	47
// #  Black
// #  	30	40
// #  Red
// #  	31	41
// #  Green
// #  	32	42
// #  Yellow
// #  	33	43
// #  Blue
// #  	34	44
// #  Magenta
// #  	35	45
// #  Cyan
// #  	36	46
// #  White
// #  	37	47
// #  
// 

func get_base() (map[string]string, error) {
//#the bash script must output the var as 
//#foo: value
//#foo: value 
//#..and the ruby script will parse that

    //#_hostname = "";  #intentionally empty unless we need to be explicit.

    //#TODO: move all this out , by using /usr/bin/env sh 
    //#for the base_setup.sh to know/use the path etc .

    //my_filename = $0

    prog, err := os.Executable() //NOTE: incorrect when 'go run ..'
    if err != nil {
        Errorln("os Executeable error")
        return nil , err 
    }

    Infoln(" prog = " + prog ) 


//CAUTION: linux 'dirname' allows a -z to append a NULL char and NOT a \n 
//but no option in OpenBSD
    cmd := exec.Command("dirname", prog)
    runtime_path, err := run_command(cmd) 
    if err != nil {
        Errorln("err calling dirname" ) 
    }

    
//    #strip out \n 
    runtime_path_trim := strings.Trim(  runtime_path , "\n") 
    Debugln("runtime_path: '" + runtime_path_trim +  "' " ) 

    
//    #-h $hostname 
//    #above switch NOT used. unless needed in future. 
    cmd_base_setup := exec.Command( runtime_path_trim +  "/base_setup.sh")
    
    Infoln("after exec.Command" ) 
    base_vars, err := run_command(cmd_base_setup)
    if err != nil {
        Errorln( fmt.Sprintf("base_setup.sh failed: %v" , err) )
    }

    fmt.Println("base_vars:" + base_vars) 

    fields := make(map[string]string) 

//TODO
    lines := strings.Split(base_vars, "\n") 
//     test := 
// `eee: xxxX
// SSS: KKKKKKK
// zzz: ooppp`; 
// 
//lines := strings.Split(test, "\n") 


    //NOTE: TODO regex requests a space after the : 
    // dbl check this is understood in the base-setup !!
    re, err := regexp.Compile(`^(.*?): (.*)$`) 
    if err != nil {
        Errorln("regex did not compile!") 
    }


    for i := 0; i < len(lines) ; i++ { 
//    fields = {} # or Hash.new
//    for line in base_vars.lines
        Debugln( "raw line : '" + lines[i] + "'" )
        res := re.FindStringSubmatch(lines[i])
        if len(res) != 3 {
            Errorln("regex array wrong size!") 
            continue
        }
        key := res[1] 
        val := res[2]
        Infoln("key = " + key, "underline" )
        Infoln("val = " + val, "bold" )
        fields[key] = val

    }

    Debugln( "variables from base_setup.sh..." ) 

    for k, v := range fields {
        Infoln( "field item key = " + k  + " , val = " + v )
    } 

    return fields, nil

}

func setup_logfile_dir(logfile_dir string) (bool, error) {
    Debugln( "setup_logfile_dir:  logfile_dir: '" + logfile_dir + "' " ) 
    //##CAUTION: OpenBSD does NOT have '-v' args for mkdir !!!!
    cmd := exec.Command( "mkdir", "-p", "logfile_dir") 

    res, err := run_command(cmd)
    if err != nil {
        //return false, errors.New( "mkdir command failed to run!" )
        return false, err 
    }
    
    Infoln("mkdir response: " + res ) 
    return true, nil
}


func clean_backup_dir(backup_dir string) (string, error) { 
//#setup and clean out backup dir for next processing...
    Debugln( "clean_backup_dir: backup_dir: " + backup_dir )

    if backup_dir == "/" {
        err :=  "backup_dir is root! Terminating now."
        Errorln(err)
        return "", errors.New(err)
    }

    if backup_dir[:4] != "/tmp" {
        err := "backup_dir does not start with /tmp" 
        return "", errors.New(err) 
    } 

    cmd := exec.Command("rm",  "-rf" ,  backup_dir)
    res, err := run_command(cmd) 
    if err != nil {
        err := "failed to run rm call " 
        return "", errors.New(err) 
    }

    Debugln( "clean_backup_dir: remove dir result : " + res ) 

    //CAUTION: OpenBSD doesnt have -v args for mkdir
    cmd_mkdir := exec.Command( "mkdir" , "-p" , backup_dir )
    res_mkdir, err := run_command(cmd_mkdir) 
    if err != nil { 
        Errorln("mkdir call failed!") 
        return "" , err 
    }

    Debugln("result : " + res_mkdir) 

    return res + "|" + res_mkdir , nil
}


 
func scan_source(path_dir string) (bool, error) { 
//create hashtable for the filesystem structure to then do a acl/mode comparision against .
    Debugln( "") 
    Debugln( "scan_source, starting for path <DIR> '" + path_dir + "'...", "bold")

    //reset the source map! 
    FileSourceMap = make(map[string]FileData )

    if _, err := scan_source_dir(path_dir, 0); err != nil {
        Errorln( "scan_source_dir call error!") 
        return false, err 
    }

   if _, err := show_prelim(false); err != nil { 
       Errorln( "show_prelim call error!") 
       return false, err 
   }

    return true, nil
}
 
 
func get_parent_perms(key_path string) (FileData, string, error) { 
//this filepath does NOT exist in the XML Treepath, so do up a level and get the default values. 
    Debugln( "\tkey_path: " +  key_path ) 

    last_dir_pos := strings.LastIndex( key_path , "/") 
    if last_dir_pos == -1 {
       return FileData{} , "", errors.New("/ char not found in key_path param") 
    }

    last_dir := key_path[0: last_dir_pos]

    Debugln( "\tlast_dir: " + last_dir)

    if f, has := FileMap[last_dir]; has {
        return f, last_dir, nil  
    } else { 
        err := "There is no key in the XML spec tree for '" + last_dir + "'\n\tAdjust XML spec or similar"
        Errorln(err) 
        return FileData{}, last_dir , errors.New(err) 
    }

    //TODO: just check if the this file_mode to file_mode mapping is corect 
}

 
func get_mode(file string) (string, error) { 
//do a file stat to get the Mode. 
//the perl chmod NEEDS an octal value input! 
//fyi: at THIS stage, it seems the result is bitmasked and output for the decimal output etc 
//but please note the octal printout format AND the bitwise mask 

    fi, err := os.Lstat(file)
	if err != nil {
        Errorln("could not stat file: " + file ) 
        return "", err
	}

    i_perm := fi.Mode().Perm()
    s_perm := fmt.Sprintf("%#o", i_perm ) // 0400, 0777, etc.

    Debugln( "s_perm : "  + s_perm ) 
    //seems same as above
    //Debugln( "Alt value " +   fmt.Sprintf("%04o", i_perm  & 07777) ) 


    return s_perm, nil

}

 
func scan_source_dir(cur_dir string, level int) (bool, error) {
//recusive scan into filesystem sourcedir to create hashmap of files and dirs
//to crossref with xml trees version 
    Debugln( "" ) 
    Debugln("scan_source_dir call...")
    Debugln("\tsourcedir (GLOBAL):" + Sourcedir  ) 
    Debugln("\tcur_dir: " + cur_dir )  

    full_dir := Sourcedir + cur_dir
    Debugln("\tfull_dir (joined): " + full_dir) 

    //TODO test the exact key's value against the FileMap -always!
    //With the Ruby version and it's xml-simple routine, it was the case that the
    // trialing lash HAD to be suffixed to get it in sync with the FileMap hashmap!!
    //hash_key := cur_dir + "/" 
    hash_key := cur_dir
    Debugln ( "\tfileSourceMap <DIR> key insert : " + hash_key , "bold")

    fm, err := get_mode(full_dir) 
    if err != nil {
        return false, err
    }

    fd := new(FileData) 
    fd.file_level = level 
    fd.file_type = "D" 
    fd.file_mode = fm 

    FileSourceMap[ hash_key ] = *fd

    Debugln ( "\tEach file in Dir : '" + full_dir + "'..."  ) 

    if files, err := os.ReadDir(full_dir); err != nil {
        Errorln("Cannot read dir: " + full_dir ) 
        return false, err
    }else { 
        for _, file := range files {     
            if file.Name() == "." || file.Name() == ".." {
                //Sat 12 Nov 2022 12:46:51 - does seem true
                Debugln("Yes, Glang does read . / .. entry! ", "bold","underline")
                continue 
            }

            Debugln( "")
            Debugln("\t\t\tfile: " + file.Name() ) 

            full_name := full_dir + "/" + file.Name()
            Debugln ( "\t\t\tfull_name: '" + full_name + "' "  ) 

            if file.IsDir() { 

                next_dir := cur_dir + "/" + file.Name()
                Debugln ( "\t\t\t\t<DIR>: next_dir: " + next_dir , "bold") 

                if _ , err := scan_source_dir(next_dir , level+1); err != nil {
                    Errorln ( "\t\t\t\tscan_source_dir error! for dir: " + next_dir )
                    return false, err 
                }
            } else if file.Type().IsRegular() { 

                hash_key := cur_dir + "/" + file.Name()  
                Debugln ( "\t\t\tfileSourceMap <FILE> insert, hash_key : '" + hash_key + "' ", "bold") 
                fm, err := get_mode(full_name)
                if err != nil {
                    return false, err
                }
                fd := new(FileData) 
                fd.file_type = "F" 
                fd.file_mode = fm
                FileSourceMap[hash_key] = *fd  

            }else {
                Errorln( "\t\t\tscanning filesystem: entry not a dir or file!" + file.Name())
                return false, err
            } 
        }//end for loop
    }//end is valid ReadDir. 

    return true, nil
}

// 
// def scan_tree_simple( # {{{
//         prev_path, 
//         obj_cur_dir, 
//         level, 
//         parent_default_file_user="", 
//         parent_default_file_group="", 
//         parent_default_file_mode="") 
// #take in a parsed HASH obj using the gem xmlsimple  
// #navigate its Hash tree and map over to the original source structure 
// 
//     #walk thru array[] of Hash items for the <directory> xml element...
//     #i think only ONE hash item it seems with either dir/file entries 
//     info ""
//     info "scan_tree_simple Starting..."
//     info "\tprev_path: " + prev_path , true
//     info "\tlevel: " + level.to_s
//     info "\tparent_default_file_mode: " + parent_default_file_mode 
//     info "\tparent_default_file_user: " + parent_default_file_user
//     info "\tparent_default_file_group: " + parent_default_file_group
// 
//     unless obj_cur_dir.is_a?(Array) then 
//         error ("\tinput dir for scanning NOT an array!\t level: #{level}")
//         return false
//     end
// 
//     for h in obj_cur_dir 
// 
//         unless h.is_a?(Hash) then 
//             error( "\titem inside directory array NOT a hash! class type: #{ h.class}  ")
//             return false 
//         end
// 
//         cur_path = h['name'] 
//         info "\tcur_path: #{cur_path}", true
// 
//         #either it's own settings or go to the parents value . so it trickles down ...  
//         default_file_mode = h['default_file_mode'] ||= parent_default_file_mode
//         default_file_user = h[ 'default_file_user' ] ||= parent_default_file_user
//         default_file_group = h['default_file_group'] ||= parent_default_file_group
// 
//         info( "\tdefault_file_mode = #{default_file_mode}")
//         info( "\tdefault_file_user = #{default_file_user}")
//         info( "\tdefault_file_group  = #{default_file_group}")
// 
//         joined_path = prev_path + cur_path + "/" 
//         debug("\tnew joined_path: " + joined_path ) 
//         $fileMap[joined_path] = file_data( level, "d", h['user'], h['group'] , h['mode'] , 
//                                 default_file_user, default_file_group, default_file_mode)
//         debug "fileMap key insert <DIR> '" + joined_path + "'", true
//         #the list of files ONLY. directories have their OWN array...
//         if h.has_key?('file') then 
//             file_list = h['file'] 
//             for f in file_list 
//                 unless f.is_a?(Hash) then 
//                     error( "\titem inside file array NOT a hash! class type: #{f.class}  ")
//                     return false 
//                 end
//                 
//                 hash_key = joined_path + f[ "name" ] 
//                 info "\tfileMap key insert <FILE> '" + hash_key + "' " , true
//                 $fileMap[hash_key] = file_data( level, 'f' , f[ 'user' ] , f[ 'group' ] , f[ 'mode' ] ,
//                                          default_file_user , default_file_group , default_file_mode )
//             end
//         end
// 
//         if h.has_key?('directory') then 
//             next_child_dirs = h['directory'] 
//             info "\trecursing into next dir..."
//             result = scan_tree_simple(joined_path, next_child_dirs, level+1, default_file_user, default_file_group, default_file_mode)
//             unless result 
//                 error("scan_tree_simple returned false, under : " + joined_path)
//                 return false
//             end 
//         end
// 
//     end 
// 
//     return true
// 
// end # }}}
// 
// def empty_or_nil?(p)# {{{
//     if p.nil? then 
//        return true 
//     end 
//     #TODO: assumes empty prop; test for Int/num etc 
//     if p.empty? then 
//         return true
//     end
// 
//     return false
// end# }}}
// 
// def simple_copy(path_dir="", delete=false, user=nil, group=nil )# {{{
//     #simple rsync version just for default-website for e.g , no xml tree etc 
//     #and signular file transfer 
//    
//     source = $sourcedir + path_dir
//     debug "simple_copy: source: '#{ source }' ", true
// 
//     target = $TEST_PREFIX + path_dir 
//     debug "simple_copy: target: '#{target}' " , true
// 
//     if File.exist?(source) == false then 
//         error( "\tsimple_copy: Terminating: '#{ source }' does not exist!")
//         return false 
//     end 
// 
//     if Dir.exist?(source) then 
//         # Add the slash to start copying the contents that follows the end dir and NOT the dir itself
//         source += "/"
//         #CAUTION: OpenBSD does not do -v on mkdir!!
//         mkdir = "mkdir -p #{target}"
//         debug("\tcommand call: '#{mkdir}'") 
//         res = %x( #{mkdir} )
//         debug( "\tsimple_copy: mkdir -p #{target}  result: '#{res}' ")
//     end 
// 
//     logfile_part = path_dir
//     logfile_part = logfile_part.gsub(/\// , "_") 
// 
//     rsync_dryrun = _get_dry_run() ? " --dry-run" : "" 
//     if empty_or_nil?(user) || empty_or_nil?(group) then 
//         user ||= ""
//         group ||= ""
//         debug "warning!!! user or group nil/empty" 
//         debug "\t user: " + user.to_s 
//         debug "\t group: " + group.to_s
//     end 
//     rsync_chown = ( empty_or_nil?(user) || empty_or_nil?(group) ) ? "" : " --chown #{user}:#{group}"
//     rsync_delete = delete ? " --delete" : "" 
//     rsync_backup = " --backup --backup-dir=#{$backupdir}#{path_dir}"
//     rsync_logfile = " --log-file=#{$logfiledir}/#{ logfile_part }_#{ Time.now.to_i }.log"
//     rsync_switches = "#{rsync_dryrun} -v -a --human-readable#{rsync_delete}#{rsync_chown}#{rsync_backup}#{rsync_logfile}"
//     rsync_call = "rsync#{rsync_switches} #{source} #{target} "
//     
//     debug( "rsync call to be run: '#{rsync_call}' ")
//     result = %x( #{rsync_call} )
//     debug "simple_copy: rsync result: #{result} "
//     return true
// 
// end# }}}
// 
func get_webowner() string  { 
    p := get_platform()
    if strings.Contains(p, "Alpine") {
        return "apache" 
    } else if strings.Contains(p, "OpenBSD") {
        return "www"
    }else {
        return "http"
    } 
}

 
func map_copy(path_dir string, delete_outsiders bool) (bool, error) {
//  open a xml tree spec to get mode/user/group etc 
//  recurse into all directory elements to get all file elements etc 
//  populate the hash tree with the full file path for easy lookup 
//  pass over to copysourcefiles with delete param for rsync to decide if to rm extra files NOT in source dir.  

    if _, err := os.Stat( path_dir ); err != nil {
        Errorln( fmt.Sprintf("path_dir: %v does not exist!", path_dir ))
        Errorln( "error: " + err.Error()) 
        return false, err
    }

    //TODO: turn into params , 
    //TODO  CHECK IF RUST CODE OR OTHER IS NOT AFFECTED BY --NOT-- CURRENTLY FLUSHING THESE TWO VARS!
    FileMap = make(map[string]FileData)
    FileSourceMap = make(map[string]FileData)
    fmt.Println("") 
    Infoln("Starting map_copy: '" + path_dir + "'", "bold") 
    Infoln("\tclearing both hash maps : fileMap and fileSourceMap" )

    
    // replace / . with _ chars 
    // the path will also have "." as is /logrotate.d/
    re, err := regexp.Compile(`[\/\.]`) 
    if err != nil {
        Errorln("file_part regex did not compile!") 
        return false, err 
    }

    file_part := re.ReplaceAllString(path_dir,"_") 
    Debugln( "\tfile_part: '" + file_part + "' " )

    file_name := fmt.Sprintf("%v/base_TREE_SPECS/spec%v.xml", Configdir, file_part)

    Infoln( fmt.Sprintf( "XML Spec Tree: '%v'", file_name))


    if _ , err := os.Stat( file_name ); err != nil {
            Errorln( fmt.Sprintf("file_name: %v does not exist!", file_name ))
            return false, err
    }

    if _, err := scan_tree_firehose(file_name); err != nil {
        Errorln("Error: " + err.Error() )
        return false, err
    }


    //now scan source file dir created hashtable. 
    //recusrse into real build directory and cross-ref the mode/user/group from the hashtable. 
    if _ , err := scan_source(path_dir); err != nil {
         Errorln( "\tscan_source call error!" ) 
         return false, err
    }

    if _, err := copy_source_files(path_dir, delete_outsiders ); err != nil {
         Errorln( "\tfailed result from copy_source_files: '" + path_dir + "' ")
         return false, err
    }

    return true , nil 
}


func explode_path_stack(p []string ) string {
    return strings.Join(p, "/") 
}

func (f *FileData ) extract_attrs(attrs []xml.Attr, default_perms_stack []FileData,  current_level int) { 
//cycle thru the Attribs and assign only when matched 
//when not found AND a default record is found on the Stack, get that one 


    f.file_level = current_level 
    //get last item on the stack. 
    var has_dp bool 
    var dp FileData 
    Infoln( fmt.Sprintf("len default_perms_stack = %v ", len(default_perms_stack) ))
    if len(default_perms_stack) > 0 {
        has_dp = true 
        dp = default_perms_stack[ len(default_perms_stack) - 1 ]
    }

    def_fm_found := false 
    def_fu_found := false 
    def_fg_found := false 
    
    for _, attr := range attrs {
        v := attr.Value
        switch attr.Name.Local {
            case "name":
                f.node = v

            case "mode":
                f.file_mode = v

            case "user":
                f.file_user = v

            case "group":
                f.file_group = v

            case "default_file_mode":
                def_fm_found = true
                f.default_file_mode = v

            case "default_file_user": 
                def_fu_found = true
                f.default_file_user = v

            case "default_file_group":
                def_fg_found = true
                f.default_file_group = v
        }
    }
    //when a higher-up record contains default perms
    //and the default attrs were not found on this element..
    if has_dp && !def_fm_found {
        f.default_file_mode = dp.default_file_mode 
    }
    if has_dp && !def_fu_found {
        f.default_file_user = dp.default_file_user
    }
    if has_dp && !def_fg_found {
        f.default_file_group = dp.default_file_group
    }
}

func scan_tree_firehose(dir_path string) (bool, error) { 
//firehose model: readonly forward only!
//scan all the tags in a continuous loop until tags are read and EOF is hit. 

    var current_level int = -1 // neg-zero as 'tree' root element not used.
    var path_stack []string
    var default_perms_stack []FileData

    var file *os.File
    var file_err error

    if file, file_err = os.Open(dir_path) ; file_err != nil { 
        Errorln( "cannot open file: " + dir_path ) 
        return false, file_err
    }

    dec := xml.NewDecoder(file)

    for { 
        tok, err := dec.Token()
        if err == io.EOF {
            break
        } else if err != nil {
            Errorln("token error: " + err.Error() ) 
            return false,err
            //break 
            //fmt.Fprintf(os.Stderr, "xmlselect: %v\n", err)
            //os.Exit(1)
        }

        switch tok := tok.(type) {
            case xml.StartElement:

                if tok.Name.Local == "tree" {
                    //skip the tree wrapper element
                    continue
                }

                fd := new(FileData) //new temp record.

                if tok.Name.Local == "directory" {
                    current_level++ //push the current level 
                    fd.file_type = "D"
                } else {
                    fd.file_type = "F"
                }


                fd.extract_attrs(tok.Attr, default_perms_stack, current_level )

                if fd.file_type == "D" { 
                    default_perms_stack = append( default_perms_stack , *fd ) 
                }

                path_stack = append(path_stack, fd.node) // push

                full_name := explode_path_stack(path_stack)

                FileMap[full_name] = *fd

            case xml.EndElement:
                if tok.Name.Local == "tree" {
                    continue
                }
                if tok.Name.Local == "directory" {
                    current_level-- //pop off  leaving dir
                    default_perms_stack = default_perms_stack[:len(default_perms_stack)-1] 
                }

                path_stack = path_stack[:len(path_stack)-1] // pop
        }
    }
    
     Debugln("")
     Debugln("FileMap entries!...", "bold") 
     for k, v := range(FileMap) {
         Debugln( fmt.Sprintf("Lv: %v Type:%v Key: '%v'", v.file_level, v.file_type, k ))
    }
//     for k, v := range(FileMap) {
//         fmt.Printf("\nFileMap: '%v'", k)
//         fmt.Printf("\n\t T=%v L=%v m=%v, u=%v g=%v du=%v dg=%v dm=%v ", 
//             v.file_type, 
//             v.file_level, 
//             v.file_mode, 
//             v.file_user, 
//             v.file_group, 
//             v.default_file_user, 
//             v.default_file_group, 
//             v.default_file_mode)
//     }
// 
    return true, nil

}

func get_etc_user() (map[string]int, error) { 
    if r, err := get_etc_secfile(true); err != nil {
        return nil, err 
    } else {
        return r, nil
    }
}
func get_etc_group() (map[string]int, error) { 
    if r, err := get_etc_secfile(false); err != nil {
        return nil, err 
    } else {
        return r, nil
    }
}

func get_etc_secfile(getuserfile bool) (map[string]int, error) {
//dont use directly
//parse the /etc/group file and return hashmap 
//or 
//parse the /etc/passwd file and return hashmap 

    var filepath string 
    if getuserfile {
        filepath = "/etc/passwd"
    }else{
        filepath = "/etc/group"
    }


    if  b_arr, err := os.ReadFile(filepath); err != nil {
        Errorln("cannot open the file: '" + filepath + "' !") 
        return nil, err
    } else {
        str := string(b_arr[:])
        Debugln("file data...") 
        Debugln(str)

        entries := make(map[string]int) 
        lines := strings.Split(str, "\n") 
        for _ , line := range lines { 
            Debugln("line=" + line) 
            cols := strings.Split(line, ":") 

            if len(cols) < 3 {
                //probably the trailing \n 
                continue 
            }

            kname := cols[0] 
            id_str := cols[2]
            id , id_err := strconv.ParseInt(id_str , 10, 32  ) 
            if id_err != nil {
                Errorln("id parse error") 
                Errorln(id_err.Error())
                return nil, err
            }

            Infoln( fmt.Sprintf("key name='%v' , id='%v' ", kname, id)) 
            entries[kname] = int(id)
        }

        //for k, x := range groups {
        //    fmt.Printf("\n G=%v , ID=%v", k, x)
        //}

        return entries, nil

    }

    //if all goes well. shoudnt get here
    return nil, errors.New("get_etc_secfile default fail") 
}

func copy_source_files(path_dir string, delete_outsiders bool) (bool, error) { 
//re-chmods the files/dirs that are in the preset TMP dir --NOT the target files 
//re-chowns the '' '' ''
//THEN rsync that dir structure across.
    Debugln("")
    Debugln("copy_source_files::Copying FileSourceMap data...") 
    Debugln ("fileSourceMap files...") 
    for k,v := range FileSourceMap {

        source_file := Sourcedir + k
        Infoln("source_file: " + source_file) 

        Debugln(fmt.Sprintf("Copying: '%v' \n\t(L:%v) key:'%v' (%v) user:%v group:%v  mode:%v\n", 
            source_file, 
            v.file_level,
            k,
            v.file_type,
            v.file_user,
            v.file_group, 
            v.file_mode ))
              
        //CAUTION!!! chmod NEEDS OCTAL value! not string, or decimal!!!
        m := v.file_mode
        if o_mode, err := strconv.ParseUint(m,8,32); err != nil {
            Errorln("file_mode didnot parse okay!") 
            return false, err
        }else{
            var fmode fs.FileMode
            fmode = fs.FileMode(o_mode)
            Debugln( fmt.Sprintf("mode (oct): %v ", fmode ))
            //Debugln("Stubbed CHMOD!") 
            if err := os.Chmod(source_file, fmode) ; err != nil{
                Errorln( "File: '" + source_file + "' did not chmod" )
                return false, err
            }
        }

        glist, err := get_etc_group()
        if err != nil { 
            return false, err 
        }
        ulist, err := get_etc_user()
        if err != nil { 
            return false, err 
        }

        g := v.file_group 
        u := v.file_user 

        var gid int 
        var has_gid bool 
        if gid, has_gid = glist[g]; !has_gid {
            g_notfound_err := errors.New("group name not found:" + g)
            return false , g_notfound_err
        }

        var uid int 
        var has_uid bool 
        if uid, has_uid = ulist[u]; !has_uid {
            u_notfound_err := errors.New("user name not found:" + g)
            return false , u_notfound_err
        }
        

        Debugln( fmt.Sprintf( "user:%v, uid:%v", u, uid ))
        Debugln( fmt.Sprintf( "group:%v, gid:%v", g, gid ))

        //Debugln("STUBBED CHOWN!!") 
         if err := os.Chown(source_file, uid, gid) ; err != nil{
             Errorln( "File: '" + source_file + "' did not chown correctly." )
             return false, err
         }

    }

    //prefix normally /home/foo/Downloads/perl_test to safeguard against overcopy.
    target_dir := TEST_PREFIX + path_dir
    Debugln("target_dir: " + target_dir)
     
    //CAUTION: OpenBSD does not do -v for mkdir 
    cmd_mkdir := exec.Command("mkdir", "-p", target_dir ) 
    if r, err := run_command(cmd_mkdir); err != nil {
        Debugln("mkdir response:" + r) 
       return false, err 
    }
    time_now := time.Now().Unix()
    cmd_rsync := exec.Command("rsync") 
    var rsync_args []string 
    logfile_part := strings.Replace( path_dir, "/", "_", -1) 

    if DryRun {
        rsync_args = append(rsync_args, "--dry-run")
    }
    rsync_args = append(rsync_args, "-a" )
    rsync_args = append(rsync_args, "--human-readable" )
    rsync_args = append(rsync_args, "--verbose" )
    if delete_outsiders { 
        rsync_args = append(rsync_args, "--delete" )
    }
    rsync_args = append(rsync_args, "--backup" )
    rsync_args = append(rsync_args, "--backup-dir=" + Backupdir + path_dir  )
    rsync_args = append(rsync_args, fmt.Sprintf( "--log-file=%v/%v_%v.log", Logfiledir, logfile_part, time_now ))
    // IMPORTANT! use the trailing  '/' at end of rsync source to avoid starting at the dir, ..so to get contents of the dir.
    rsync_args = append(rsync_args, Sourcedir + path_dir + "/"  )
    rsync_args = append(rsync_args, target_dir )

    // TODO: Rust's version FAILS when extra blank space chars are between args. Dbl check here. 
    // #TODO parse the stdout response!!!
    // # this assumes it ran okay!

    cmd_rsync.Args = rsync_args 
    for _, a := range cmd_rsync.Args {
        Debugln("rsync Arg: " + a ) 
    }

    if r, err := run_command(cmd_rsync); err != nil { 
        Debugln("rsync command failure! : " + r) 
        return false, err
    } else {
        Debugln("rsync non-err return: " + r )
    }

    //all good if arrived here. 
    Infoln("rsync completed okay!")
    return true, nil
}
 
func show_prelim(this_is_re_show bool) (bool, error)  {
//show to user What will happen re file Mode, Missing etc   
//iterate the xmltree first then the filesys source tree 

    Say( "")
    Say( "====================== XML Tree spec map ==============================" ) 
    Say( "??? = File missing from XML spec master file.")
    Say( "=======================================================================")
    Say( "")

    for key , item := range FileMap {
        var alert string 
        if _, has := FileSourceMap[ key ]; has {
            alert = "   " 
        }else {
            alert = "???"
        }

        var ftype string 
        if ftype = item.file_type; ftype == "" {
            ftype = "?"
        }

        Say( fmt.Sprintf("%v %v %v:%v %v L%v %v\n" ,
            alert,
            ftype, 
            item.file_user, 
            item.file_group, 
            item.file_mode,
            item.file_level , 
            key, 
        ))
    } 

    Say("")
    Say("===================== Filesystem source map ===========================")
    Say("??? = File not listed in XML Tree spec. ")
    Say("XXX = File's mode will be overridden to match the XML file's version. ")
    Say( "      <<OVERRIDE>>  OLD --> NEW ")
    Say( "=======================================================================")
    Say( "" )

    for key, item := range FileSourceMap {
        var msg string
        var alert string
        if f, has := FileMap[key]; has  {
            //it exists in the XML treemap...
            //the FileSourceMap CANNOT really have the target user/group as it is coming from a dev machine anyway. 

            item.file_user = f.file_user
            item.file_group = f.file_group
            if f.file_mode != item.file_mode {
                alert = "XXX"
                msg = "<<OVERRIDE>> " + item.file_mode + " --> " + f.file_mode
                //RESET value to match the XML spec.
                //FileSourceMap[key].file_mode = f.file_mode
                item.file_mode = f.file_mode
            }

           FileSourceMap[key] = item  
        } else {
            //missing file: 
            //the file in the sourcemap is NOT in the XML tree spec. 
            //get last dir / go up a dir and get the default perms for that file. 
            if perms, last_dir, err := get_parent_perms(key); err != nil {
                Errorln( "Failed to get parent permissions for '" + key + "' ")
                return false, err 
            } else {
                alert="???"
                msg="**Missing** (owner dir: " + last_dir + ")"
                item.file_user = perms.file_user
                item.file_group = perms.file_group
                item.file_mode = perms.file_mode
            }//
        } 

        if len(alert) == 0 {
            alert = "   "
        }

        Say ( fmt.Sprintf("%v %v %v:%v %v L%v %v %v\n" ,
                    alert,
                    item.file_type, 
                    item.file_user ,
                    item.file_group ,  
                    item.file_mode ,
                    item.file_level , 
                    key, 
                    msg))

        //reset the value to the updated FileData struct!
        FileSourceMap[key] = item 
    } //endfor

    Debugln( fmt.Sprintf( "XML Tree Spec record count: %v", len(FileMap)))
    Debugln( fmt.Sprintf( "  File-source record count: %v", len(FileSourceMap)))

    if ForceYes {
        Say( "FORCING a Yes for all would-be user input!")
    }else{
        Say( "Considering all above, proceed with the file copy tasks? y/N")

        var answer string 
        if _ , err := fmt.Scan(&answer); err != nil {
            Errorln("Did not scan line correctly!") 
            return false, err 
        }

        Debugln( "STDIN: answer:  '" + answer + "' " )

        if answer == "y" || answer == "Y" {

            if this_is_re_show == false {
                if _, err := show_prelim(true); err != nil {
                    return false, err
                }
            }

            Say( "Answered 'Yes', Now Processing...")

        } else if answer == "N" || answer == "\n" || answer == "n" {
            Say( "Answers 'No' -Bailing out of the map_copy!")
                return false, errors.New("A 'no' answer was taken.")
        } else {
            err := "Could not understand response. Terminating now. "
            Errorln(err)
            return false, errors.New(err)
        }
    }

    return true, nil
} 

func set_globals() (bool , error) { 
//store the common variables from common shell script
//the globals from the command line should already be set. 

    recs, err := get_base()
    if err != nil {
        Debugln("set_globals: recs type: " + err.Error() ) 
        return false, err
    }
        
    if len(recs) == 0  {
        Errorln( "get_base() did not return any records!" ) 
        return false, errors.New("empty set") 
    }

    if  _ , found := recs["ERROR"]; found {
        Errorln( "Terminal ERROR record found for base_setup" ) 
        return false, errors.New("ERROR record is dependency!") 
    } 

    var ok bool 
    if Configdir, ok = recs["configdir" ]; !ok {
        return false, errors.New("configdir missing!")
    }

    if Swapdir, ok = recs["swapdir" ]; !ok {
        return false, errors.New("swapdir missing!")
    }

    if Target, ok = recs["target" ]; !ok {
        return false, errors.New("target missing!")
    }

    if Buildname, ok = recs["buildname" ]; !ok {
        return false, errors.New("buildname missing!")
    }
    if Builddir, ok = recs["build_dir" ]; !ok {
        return false, errors.New("build_dir missing!")
    }

    Sourcedir = Builddir

    Backupdir = Swapdir + "/base_backup_BUILD_" + Target
    Logfiledir = Swapdir + "/rsync_log"

    Debugln( fmt.Sprintf( "sourcedir = %v", Sourcedir ) )

    //just ignore the NULL suffix, as a Dev machine was most likely matched 
    null:="NULL"
    var target_null_test string
    if BypassTargetNull {
        target_null_test = "" 
    }else {
        target_null_test = null
    }
        
    if Target == "" || Target == target_null_test {
        err := "target is NULL or empty." + 
                "\n\tTerminating process" + 
                "\n\tProbably running on the bare metal dev machine. this is a no-no. "
        Errorln(err) 
        return false , errors.New(err) 
    }


    Debugln(fmt.Sprintf( "ARG: debug: %v", Debug ))
    Debugln(fmt.Sprintf( "ARG: run_mode = %v", RunMode))
    Debugln(fmt.Sprintf( "ARG: force_yes = %v", ForceYes))
    Debugln(fmt.Sprintf( "ARG: dry_run = %v, ", DryRun ))
    Debugln(fmt.Sprintf( "ARG: bypass_target_null = %v" , BypassTargetNull))
    Debugln(fmt.Sprintf( "TEST_PREFIX:  '%v' ", TEST_PREFIX ))

    if DryRun {
        Infoln( "Running in DRY-RUN mode for rsync, no changes saved!!!")
    }

    return true, nil

} 
 
func get_command_lines(full_csv_path string) ([][]string, error) {

    if _, err := os.Stat( full_csv_path ); err != nil {
        Errorln( fmt.Sprintf("full_csv_path: %v does not exist!", full_csv_path ))
        Errorln( "error: " + err.Error()) 
        return nil, err
    }

    var file os.File  
    var file_err error
    if file, file_err = os.Open(full_csv_path) ; file_err != nil { 
        Errorln( "cannot open file: " + full_csv_path ) 
        return nil, file_err
    }

    r := csv.NewReader(file)
	r.Comma = ','
	r.Comment = '#'
	records, err := r.ReadAll()

    if len(records) == 0 {
      Errorln( "all empty command lines!") 
      return nil, errors.New("All empty Command lines in file!") 
    } 

    for _, line := range records { 
        fmt.Println("")
        fmt.Print("Rec:: ") 
        for _, col :=range line {
            fmt.Print(col + ",") 
        }
    }
    
    return records, nil 
} 
 
func lookup_user(p string) string {
//TODO Etc call
   return p
}
 
func lookup_group(p string) string {
   //TODO Etc call
   return p
}
 
func parse_path(path string) (string, error) {
    p := strings.TrimSpace(path) 

    if p == "/" { 
        Errorln( "path is root!" ) 
        return "", errors.New("Path is root!")
    } 

  //TODO regex for url/path
    return p, nil //# or false
}

func parse_bool(_p string) (string, error) {
    p = strings.TrimSpace(_p)
    if !(p == "true" || p == "false"){ 
        return "", errors.New("value neither true or false") 
    }
    return p, nil  //return the STRING
}

func convert_bool(p string ) bool { 
    if p == "true" { 
        return true
    } 
    return false
}

func parse_user(_p string) (string, error) { 
    p = strings.TrimSpace(_p) 

    if p == "" { 
        return "", errors.New("empty user given!") 
    }

    var new_user string 
    if p == "<webowner>" {
        new_user = get_webowner()
    }else{
        new_user = p 
    }

    if ulook, err := lookup_user(new_user); err != nil { 
        return "", err 
    } else { 
        return ulook, nil 
    }

    return "", errors.New("invalid parse_user outcome.")

}

func parse_group(_p string) (string,error) {
  
    p = strings.TrimSpace(_p)
    if p == "" {
        return "", errors.New("empty group param.") 
    }

    var new_g string 

    if p == "<webowner>" {
        new_g = get_webowner()
    }else {
        new_g = p 
    }

    if look_g, err := lookup_group(new_g); err != nil { 
        return "", err
    }else {
        return look_g, nil
    }

    return "", errors.New("invalid outcome") 

}

func parse_simple_cmd(cmd []string) (Command, error) { 
//attempt simple_copy parse 
 
// s,"/var/www/html/sites/default", false, <webowner>, <webowner>
// remember! the first element was SHIFTed , so s is gone!
// 
    Debugln("cmd: + cmd") 
 
    if len(cmd) < 4 {
        return nil, "array size less than 4"
    } 
    pos_path:=0
    pos_delete:=1
    pos_user:=2
    pos_group:=3
 
    command := new(Command)
    command.c_type = "simple"

    for i, v := range cmd { 
        
        switch i {
            case pos_path;
                if param_path, err := parse_path(cmd[pos_path]); err != nil { 
                     return nil, errors.New( fmt.Sprintf( "position %v is not a valid path.",pos_path )) 
                }else {
                    command.c_path = param_path 
                }
            case pos_delete; 
                if param_del, err := parse_bool(cmd[pos_delete]); err != nil { 
                     return nil, errors.New( fmt.Sprintf( "position %v is not a valid boolean.",pos_delete )) 
                }else {
                    command.c_delete = convert_bool(param_del) //convert AFTER guard. 
                }
            case pos_user; 
                if param_user, err := parse_user(cmd[pos_user]); err != nil { 
                     return nil, errors.New( fmt.Sprintf( "position %v is not a valid user.",pos_user )) 
                }else {
                    command.c_user = param_user 
                }
            case pos_group; 
                if param_group, err := parse_group(cmd[pos_group]); err != nil { 
                     return nil, errors.New( fmt.Sprintf( "position %v is not a valid group.",pos_group )) 
                }else {
                    command.c_group = param_group 
                }

        }

    }
 
    return command, nil
 
} 
 
// def parse_mapcopy_cmd(cmd)
// 
// # m,"/etc/httpd/conf" , false
// # remember - the first element SHOULD of been shifted!
// 
//   if cmd.size < 2 then 
//     return nil, "array size less than 2"
//   end 
// 
//   pos_path=0
//   pos_delete=1
// 
//   if false == param_path = parse_path(cmd[pos_path] ||= "") then 
//     return nil, "position #{pos_path} is an invalid path." 
//   end
// 
//   if false == param_delete = parse_bool(cmd[pos_delete] ||= "") then 
//     return nil, "position #{pos_delete} is not a bool"
//   else
//     param_delete = convert_bool(param_delete) #convert after guard,
//   end
// 
//  
//   command = { 
//     "type" => "mapcopy", 
//     "path" => param_path, 
//     "delete" => param_delete, 
//   }
// 
//   return command, false
// 
// end
// 
// def parse_command(cmd)
// 
//   action = cmd.shift
//   action = action.strip
//   
//   unless action == "s" || action  == "m"
//     return false
//   end
// 
//   command = false
//   err = false
// 
//   if action == "s" then 
//     command, err = parse_simple_cmd(cmd)
//   elsif action == "m" then 
//     command, err = parse_mapcopy_cmd(cmd)
//   else 
//     err = "unknown option" 
//   end 
// 
//   return command, err
// 
// end 
// 
// 
// def create_command_list(config_dir, command_csv_file)
// #push good parsed commands to the array 
// 
//   full_path = config_dir + "/" + command_csv_file 
//   lines = get_command_lines(full_path) 
//   unless lines 
//     error "get_command_lines failed!" 
//     return nil, false
//   end 
// 
//   debug "command lines: " + lines.to_s
// 
//   delim = ","
//   command_list = Array.new
// 
//   for x in lines
//       command, err = parse_command( x.split(delim) )
//       if err then 
//         error "failed to parse command!"
//         error "\t" + err.to_s
//         return nil, false
//       end 
//       command_list.push(command)  
//   end
// 
//   return command_list, nil
// 
// end 
// 
// def process_commands(command_list)
// 
//   unless command_list.is_a?(Array) 
//     error "command_list not an Array!"
//     return -1
//   end 
// 
//   if command_list.size == 0 then 
//     error "command_list size zero!"
//     return -1
//   end 
// 
//   for c in command_list 
// 
//     result = false
// 
//     if c['type'] == "simple" then 
//       debug "COMMAND:: path: " +  c['path'] 
//       debug "COMMAND:: " + c.to_s 
//       result = simple_copy(c['path'], c['delete'], c['user'] , c['group']) 
// 
//     elsif c['type'] == "mapcopy" then 
//       debug "COMMAND:: path: " + c['path'] 
//       debug "COMMAND:: " + c.to_s
//       result = map_copy(c['path'], c['delete']) 
//     end
// 
//     unless result 
//       error "call failure with simple_copy or map_copy command!" 
//       return false
//     end 
// 
//   end #endfor
// 
//   info "processed commands okay!", true
//   return true
// end 
// 
// 
// def _get_dry_run() 
//   return $dry_run
// end 
// 
// def _get_debug() 
//   return $debug
// end 
// 
// def _get_force_yes()
//   return $force_yes
// end
// 
// 
// ###############################################################################
// #                           Logic Start...
// ###############################################################################
// 
// $VERSION = "0.1.0"
// 
// 
// mapcopy_csv_file="mapcopy_commands.csv"
// 
// terminate, $dry_run, $force_yes, $run_mode, $bypass_target_null, $debug = set_args() 
// if terminate 
//     #use at bash/sh, echo $? to show the result
//     info "terminating process (set_args)"
//     exit 1
// end
// 
// splash($VERSION)
// 
// 
// unless 0 == result = set_globals() then 
//     error "set_globals() failure" 
//     exit 2
// end
// 
// ###########TODO
// #testing the reshuffle of global vars. 
// #remove once confident. 
// 
// unless clean_backup_dir($backupdir)
//     error "clean_backup_dir() call failure" 
//     exit 3
// end 
// 
// unless 0 == x = setup_logfile_dir($logfiledir) 
//     error "setup_logfile_dir() call failure" 
//     exit 4
// end 
// 
// webowner = get_webowner()
// debug("webowner: #{webowner}") 
// 
// 
// command_list, err = create_command_list($configdir, mapcopy_csv_file) 
// unless err == nil 
//   error "create_command_list failed!" 
//   exit 5
// end 
// 
// #TODO Need to decide if another ARGV is needed to proceed-on-error 
// # or terminate of the first/any error and ignore the outstanding items in list.
// 
// unless true == result = process_commands(command_list)
//   error "process_commands failed!"
//   exit 6
// end 
// 
// 
// exit 0 #zero is unix success. 
// __END__
// 
// #################### now obselete.......
// # all these should be in the CSV file!
// # keep for posterity / reference.
// # simple_copy("/var/www/html/sites/default", false, webowner, webowner )  
// # simple_copy("/var/www/html/sites/default_http" , false, webowner, webowner )  
// # simple_copy("/var/www/html/sites/default_https" , false, webowner, webowner )  
// # 
// # map_copy("/etc/httpd/conf" , false )  
// # map_copy("/etc/apache2" , false )  
// # map_copy("/etc/postfix" , false )  
// # map_copy("/etc/postgresql" , false )  
// # map_copy("/etc/php" , false )  
// # map_copy("/etc/php8" , false )  
// # map_copy("/var/lib/postgres" , false )  
// # map_copy("/var/lib/postgresql" , false )  
// # map_copy("/root", false )  
// # map_copy("/home/vagrant", false )  
// # map_copy("/home/arch", false )  
// # map_copy("/home/alpine", false )  
// # map_copy("/etc/logrotate.d" , false )  
// # 
// # simple_copy("/etc/redis.conf" , false, 'redis', 'redis' )  
// # simple_copy("/etc/ssl_self" , false, 'root', 'wheel' )  
// # simple_copy("/etc/letsencrypt" , false, 'root', 'wheel' )  
// # ##################################################################
// 
// 
// #goodbye. 
// 
// 
// 



