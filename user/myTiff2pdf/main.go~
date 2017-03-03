package main

import (
    "fmt"
    "user/files"
    "flag"
    //"runtime"
    "io"
    "strings"
    "io/ioutil"
    "os"
    "os/exec"

)

func usage() {
    usage := "usage: MergeTiff -i inpuDir -o outputDir]"
    fmt.Println(usage)
    flag.PrintDefaults()
    os.Exit(2)
}

func exec_cmd1(cmd string, args string) error {
	
	path, err := exec.LookPath(cmd)
         
	if err != nil {
		fmt.Println(err,"lookpath")

	} else {
		argA := strings.Fields(args)

		for k, v := range argA {
			argA[k] = v
			//fmt.Println(argA[k])
		}
		_, err = exec.Command(path, argA...).Output()
	}
	return error(err)
}


func exec_cmd(cmd string) {
    // splitting head => g++ parts => rest of the command
  		parts := strings.Fields(cmd)
  	//head := parts[0]
  	parts = parts[1:len(parts)]
  	 
  	// out, err := exec.Command(head,parts...).Output()
        out,err := exec.Command("ls","*").Output()
  	if err != nil {
         fmt.Printf("%s", out)
   	 fmt.Printf("%s", err)
  	}
  	fmt.Printf("%s", out)
   
    
   
}


func main() {

var (
     inputDir string
     outputDir string
)

flag.Usage = usage
flag.StringVar(&inputDir,"i","/home/paul/part0000/Tiff","")
flag.StringVar(&outputDir,"o","/home/paul/part0000/Tiffcp","")
flag.Parse()
if len(outputDir) == 0 {
      usage()
}




/*cpath,_  := user.Current()
path     := cpath.HomeDir+"/pdf/"
*/
var ufile []string
input_exist,_ := files.Exists(inputDir)
output_exist,_ := files.Exists(outputDir)

if output_exist == false {
   if  os.MkdirAll(outputDir,0755) != nil {
      fmt.Println("Can't make output fir")
      os.Exit(1)
   }
}
var fn2 string
if input_exist == true {
       // read directory entries
	dirent,_ := ioutil.ReadDir(inputDir) 	
	//var buf []byte 
	//var filesize int64  
	for _,file := range dirent {
 	//os.fileInfo
		 
		filename := file.Name()
                fn := strings.Split(filename,".")[0]
                doctype  := strings.Split(filename,".")[1]
		if doctype != "tiff" {
                      	fmt.Println("Skip",fn)
                        break
                }    
		fn1 := fn[:len(fn)-5]
                if fn1 != fn2 {
                  fn2= fn1
                  ufile = append(ufile,fn1)      

                } 

	}
        cmd:= "tiffcp"
        cmd1 := "tiff2pdf"   
        path,_   := exec.LookPath(cmd)
        path1,_ := exec.LookPath(cmd1)
        sep := string(os.PathSeparator)
        home:= os.Getenv("HOME")
        bin := home+sep+"bin"
        cmd = bin+sep+"mytiffcp.sh"
        fmt.Println(cmd)

        fp,err := os.Create(cmd)
        if err !=nil {
            panic(err)
        }
          
        for _,v := range ufile {
           
           //fmt.Println("processing",v)  
           ofile := outputDir + sep + v 
           s1  := path+ " " + inputDir+sep+v+"_????.tiff" + " " + ofile +".tiff\n"
           _,err := io.WriteString(fp,s1)
          
           s2  :=  path1 + " "+ ofile +".tiff -o " + ofile+".pdf\n" 
           _,err = io.WriteString(fp,s2) 
           if err != nil {
              fmt.Println(err)
           }        
        }
        fp.Close()
}        
}




 

