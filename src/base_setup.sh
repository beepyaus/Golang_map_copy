#!/usr/bin/env sh

#WAS...(with! after #)
#/bin/sh

#Usage: base_setup.sh -h HOSTNAME -s /home/user/foo/path/base_setup.sh


#not to be called directly
#Used by other scripts and perl/rust/etc to get global vars 
#NOTE: devpath update if location change 

#NOTE OpenBSD using the /etc/hostname files very differently from linux, 
# nic: network interface card/virtual device 
# OpenBSD uses /etc/hostname.{nic_tag} /etc/hostname.{nic_tag_wifi}
# hence the hostname call and NOT an environment variable
#docker uses an explicit hostname parameter


#####Globals########
hostname=
target=
configdir=
base_rootdir=
target_rootdir=
build_dir=
buildname=
swapdir= 
platform=
webuser=
##################

set_webuser()
{
    platform="$( uname -s )"
    echo "platform: ${platform}"

    webuser=
    if [[ "$platform" == "alpine" ]]; then
        webuser="apache"
    elif [[ "$platform" == "OpenBSD" ]]; then
        webuser="www"
    else 
        #maybe Linux (dev machine other)
        webuser="http"
    fi
}

base_setup_get_fields() 
{
    #echo "my param: ::0 = $0" 
    #echo "my param: ::1 = $1" 
    #$1 == hostname 

    #used when NOT called via Docker -OS cmd call.
    default_hostname="$(hostname)"

    #this_dir="$1"

    #operating system hostname
    hostname="$1"

    #full name of script, as BASH_SOURCE[0] is invalid 'sh'/ksh value. 
    self=

    hostname="${hostname:-$default_hostname}"

    if [[ -z "$hostname" ]]; then 
        echo "ERROR: hostname arg empty!" 
        exit; 
    fi

    # this line below: only for bash - NOT sh/ksh etc
    # thisdir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

    target=
    if [[ "$hostname" == "vagrant" ]]; then 
        target="VAGRANT"
    elif [[ "$hostname" == "alpinevagrant" ]]; then 
        target="VAGRANT"
    elif [[ "$hostname" == "docker" ]]; then 
        target="DOCKER"
    elif [[ "$hostname" == "thinklin" ]]; then 
        target="NULL"
    elif [[ "$hostname" == "maclin" ]]; then 
        target="NULL"
    elif [[ "$hostname" == "openjobi.localhost" ]]; then 
        target="VBOX_OPENBSD"
    elif [[ "$hostname" == "vultral" ]]; then 
        #vultra is the Vultr VM 
        #PROD/LIVE -but- currently in testing Fri 02 Jul 2021 18:04:42
        target="PROD_VUL"
    else 
        #we default to production so if the above matches did not work,
        #the importance of the production enviroment is not compremised.
        #and the file updates should get to their destination. 
        target="PROD"
    fi

    #  thisdir="$( dirname $self )";
    #  if [[ $thisdir == "" ]]; then 
    #      echo "ERROR: empty 'thisdir' variable!"
    #      exit;
    #  fi 

    if [[ -z "$SYNC_PATH" ]]; then 
        echo "ERROR: \$SYNC_PATH empty! Terminating process."
        return 1; 
    fi

    configdir="${SYNC_PATH}/assets/config"
    configdir="$( cd $configdir && pwd )"
    buildname="base_BUILD_${target}"
    base_rootdir="$( cd "${configdir}/base" && pwd )"
    target_rootdir=${base_rootdir}_${target}

    #copy and dump/merge this dir to be deployed to the target machine
    swapdir="/tmp/deployswap/build"

    #build straight into the swap/tmp dir!!
    # keeps the dev dir structure clean for git etc 
    # have no re-chown in the DEV project! 
    build_dir="${swapdir}/${buildname}"

    #IMPORTANT!! echo vars for perl call to access in per line 
    #Output format is key:value 
    #Do NOT have any other echo/output lines in this file!
    echo "hostname: $hostname"
    echo "target: $target"
    echo "configdir: $configdir"
    echo "base_rootdir: $base_rootdir"
    echo "target_rootdir: $target_rootdir"
    echo "build_dir: $build_dir"
    echo "buildname: $buildname"
    echo "swapdir: $swapdir"
    echo "TEST_DOLLAR_ZERO: $0"
}

#either 'sourcing' the file via another bash/sh file OR 
#Rust/Ruby/Perl script accessing it via direct OS system command call. 
#DO NOT enable this variable in THIS file -only calling file. 
if [[ -z "$FUNCTION_CALL" ]]; then 

    my_host_name=

    #echo "debug: using getopts section"

    #ONLY send the explicit -h param when it is DOCKER!
    #while getopts "h:s:" name
    while getopts "h:" name
    do
        case "$name" in
            h)  my_host_name="${OPTARG}" ;;
            #s)	my_self="$OPTARG" ;;
            #?)	echo "Usage: -h my_host_name -s my_full_script_path "; exit 2 ;;
        esac
    done
    shift $(($OPTIND - 1))

    # maybe not needed. 
    # if [[ -n "$@" ]]; then 
    #     echo "ERROR: arguments outside scope -> " "$@"
    # fi
    # 
    # #this self hopfully not needed. 
    # if [[ -z "$self" ]]; then 
    #   echo "ERROR: s (self) arg not set!" 
    #   exit; 
    # fi

  base_setup_get_fields "$my_host_name"

fi 


