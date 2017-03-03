package originalpdf

import (
 	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"user/files"
	"encoding/json"
	 
)

func CreateFromFolder(inputDir string,outputdir string) (error) {
	
	var ( 
	  err error
	  level1_dirent []os.FileInfo
	  level2_dirent []os.FileInfo
	  level3_dirent []os.FileInfo
	)

	if level1_dirent,err = ioutil.ReadDir(inputDir); err != nil {
		fmt.Println("Error Reading",inputDir,err)
		return err

 	} 
	outputDir := path.Join(outputdir,"Pdf")
	contDir   := path.Join(outputdir,"Container")
	if  !files.Exist(outputDir) {

		if  os.MkdirAll(outputDir,0755) != nil{
			fmt.Println(err)
			os.Exit(2)
		}

	}

	if  !files.Exist(contDir) {

		if  os.MkdirAll(contDir,0755) != nil{
			fmt.Println(err)
			os.Exit(2)
		}

	}

		
	dirname := strings.Split(inputDir,"/")
	pdf_s := dirname[len(dirname)-1]
	
	usermd := map[string]string{} 			
  	for _,level2_dir := range level1_dirent {
		level2_path := path.Join(inputDir,level2_dir.Name())	
		if level2_dirent,err = ioutil.ReadDir(level2_path);err !=nil {
			fmt.Println("Error reading",level2_path,err)
			return err
		}
		
		for _,level3_dir := range level2_dirent {
			level3_path:= path.Join(level2_path,level3_dir.Name())	
			if level3_dirent,err = ioutil.ReadDir(level3_path);err !=nil {
				fmt.Println("Error reading",level3_path,err)
				return err
			}
			for _,level4_dir := range level3_dirent {
				level4_path := path.Join(level3_path,level4_dir.Name())
				level4_name := 	level4_dir.Name()
				name := strings.Split(level4_name,".")[0]
				ftype:= strings.Split(level4_name,".")[1]	
				if ftype ==  "pdf" {	
					kc := 	level3_dir.Name()+"1"	
					date := level2_dir.Name()
					pdf_fn := pdf_s+date+name+kc+ "."+ftype 	
					ct_fn := pdf_s+date+name+kc+ ".json"
						
					pdf_file_in := level4_path 
					pdf_file_out := path.Join(outputDir,pdf_fn)
					fmt.Println("copying from:",pdf_file_in,"to", pdf_file_out)

					usermd["page_number"]="0000"
					usermd["pub_office"]=pdf_s
					usermd["o_pub"]= pdf_s
					usermd["kc"]= kc
					usermd["Date_drawup"]= date
					usermd["Pub_date"]=date
					usermd["doc_id"] = date+name 
	
					ct_file_out := path.Join(contDir,ct_fn) 
					if bytes,err:= ioutil.ReadFile(pdf_file_in); err == nil {

						if ioutil.WriteFile(pdf_file_out,bytes,0644) != nil {
							fmt.Println(err, "Writing",pdf_file_out)
							os.Exit(2)					
						}
						usermdB, _ := json.Marshal(usermd)
						if  ioutil.WriteFile(ct_file_out, usermdB, 0644) != nil {
							fmt.Println(err, "Writing",ct_file_out)
							os.Exit(2)
						}
						 						
					 
					}
					
					
				}
			}
							 
		}
				
	} 
return err
}
 



