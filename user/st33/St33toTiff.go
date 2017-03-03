package main

import (
  "fmt"
  "io"
  "flag"
  "os"
  "io/ioutil"
  "strconv"
  "bytes"
  "strings" 
  "user/ebc2asc" 
  "encoding/binary"
  "encoding/json"
  //"user/tiff" 
  
)

type Configuration struct {
   Input_directory  string
   Input_files      []string
   Output_tiff      string
   Output_json      string
}

const (
	leHeader = "II\x2A\x00" // Header for little-endian files.
	beHeader = "MM\x00\x2A" // Header for big-endian files.

	ifdLen = 10 // Length of an IFD entry in bytes.
        
)
 
const (
	dtByte     = 1
	dtASCII    = 2
	dtShort    = 3
	dtLong     = 4
	dtRational = 5
)

// The length of one instance of each data type in bytes.
var lengths = [...]uint32{0, 1, 1, 2, 4, 8}

// Tags (see p. 28-41 of the spec).
const (
	tImageWidth                = 256
	tImageLength               = 257
	tBitsPerSample             = 258
	tCompression               = 259
	tPhotometricInterpretation = 262

	
        tStripOffsets    = 273
        tOrientation    = 274
	tSamplesPerPixel = 277
	tRowsPerStrip    = 278
	tStripByteCounts = 279

	tXResolution    = 282
	tYResolution    = 283
	tResolutionUnit = 296

	tPredictor    = 317
	tColorMap     = 320
	tExtraSamples = 338
	tSampleFormat = 339
)

const (
	cNone       = 1
	cCCITT      = 2
	cG3         = 3 // Group 3 Fax.
	cG4         = 4 // Group 4 Fax.
	cLZW        = 5
	cJPEGOld    = 6 // Superseded by cJPEG.
	cJPEG       = 7
	cDeflate    = 8 // zlib compression.
	cPackBits   = 32773
	cDeflateOld = 32946 // Superseded by cDeflate.
)

var configfile  string 


var Little = binary.LittleEndian 
var Big = binary.BigEndian 
var enc = Big
var ifdOffset = 8
var Tiff_header_size = 160  // 0xA0
var xoffset = 134
var yoffset = 142

func usage() {
    usage := "usage: st33ToTiff -c config_file -i input_file -d input_directory"
    fmt.Println(usage)
    flag.PrintDefaults()
    os.Exit(2)
}

func getConfig(configfile string) (Configuration,error) {

  cfile,err := os.Open(configfile)
    if err != nil { panic(err) }
  decoder      := json.NewDecoder(cfile)
  configuration:= Configuration{}
  err= decoder.Decode(&configuration)
  _ = cfile.Close() 
  return configuration, err

}

func check(e error)  {
  if e != nil{
     panic(e)
  }

}

func getuint16(in []byte) ( uint16){
  out,_ := strconv.Atoi(string(in)) 
  return uint16(out)

}

func getuint32(in []byte) ( uint32){
  out,_ := strconv.Atoi(string(in)) 
  return uint32(out)

}


func getOrientation(rotation_code []byte) (uint16) {
   orientation,_ := strconv.Atoi(string(rotation_code))  
   switch orientation {
     case 1: return uint16(1)
     case 2: return uint16(6)
     case 3: return uint16(3)
     case 4: return uint16(8)
   default: return uint16(1)
   }       
}


func Extend(slice []string, element string) []string {
    n := len(slice)
    if n == cap(slice) {
        // Slice is full; must grow.
        // We double its size and add 1, so if the size is zero we still grow.
        newSlice := make([]string, len(slice), 2*len(slice)+1)
        copy(newSlice, slice)
        slice = newSlice
    }
    slice = slice[0 : n+1]
    slice[n] = element
    return slice
}


func main(){
 
var InputFile string 
var inputDir  string
 flag.Usage = usage
 flag.StringVar(&configfile,"c","","")
 flag.StringVar(&InputFile,"i","","") 
 flag.StringVar(&inputDir,"d","","")
 flag.Parse()
 if len(configfile) == 0 {
     usage()
  }
 // read configuration file
 var conf Configuration 
 conf,err := getConfig(configfile)
 if err != nil {
     fmt.Println(err)
 }
 // get Configuration parms
 
 var inputFile   []string
 
 if inputDir == "" {
   inputDir  = conf.Input_directory}

 if InputFile == "" {
    inputFile = conf.Input_files
 } else {
     inputFile = Extend(inputFile,InputFile)  }

 outputTiffDir   := conf.Output_tiff
 outputusermdDir := conf.Output_json

 

 for _,file := range inputFile {
  
  filename := inputDir+string(file)    
  fmt.Println("Processing:", filename)

  buf,err := ioutil.ReadFile(filename)
  
  if err != nil {
   fmt.Println(err)
  } else {
     l:=0
     bufl := len(buf)
     usermd :=  map[string]string{}
     for    l < bufl  {
       k:= l
       
       var ( 
           recs uint16  
           total_rec uint16
           total_length uint32
           imgl uint16 
       )
       buf1 := bytes.NewReader(buf[k+25:k+27])
       err = binary.Read(buf1, Big, &recs)
       buf1 = bytes.NewReader(buf[k+84:k+86])
       err = binary.Read(buf1, Big, &total_rec) 
       buf1 = bytes.NewReader(buf[k+214:k+218])
       err = binary.Read(buf1, Big, &total_length) 
       buf1 = bytes.NewReader(buf[k+250:k+252])
       err = binary.Read(buf1, Big, &imgl)


       st33 := ebc2asc.Ebc2asci(buf[l:l+214])
       long,_ := strconv.Atoi(string(st33[0:5]))
        
       pub :=   st33[5:7] 
       usermd["pub_office"] = string(pub)
       kc :=    st33[7:9]     
       usermd["kc"] = string(kc)
       // docnum := st33[9:17] 
       pagenum := st33[17:21]
       usermd["page_number"] = string(pagenum)
       // framenum := st33[21:25]
       //recs := byte2int(buf[k+25:k+27])
       //fmt.Println("recs",int(recs)) 
       
       docid := st33[33:45]
       usermd["doc_id"] = strings.TrimLeft(string(docid)," ")
       o_pub := st33[67:69]
       usermd["o_pub"] = string(o_pub)
       //date_drawup := st33[69:75]
       //usermd["date_drawup"]= string(date_drawup)
       //rec_stat :=  st33[75:76]
       total_pages := st33[76:80]  
       usermd["total_pages"]= string(total_pages)
       /*
       s_doc_h := st33[87:90]
       s_doc_w := st33[90:93] 
       */
       f_date_drawup := st33[93:101]
       usermd["Date_drawup"]= string(f_date_drawup)
       f_pub_date := st33[101:109]
       usermd["Pub_date"]= string(f_pub_date) 
       Biblio := string(st33[133:134])
       Claim  := string(st33[134:135])
       Drawing := string(st33[135:136])
       Amendement := string(st33[136:137])
       Description:= string(st33[137:138])
       Abstract  :=  string(st33[138:139])
       Search_report:=  string(st33[139:140])
       
       var content bytes.Buffer

       if Abstract   == "1" { content.WriteString("A")}
       if Biblio     == "1" { content.WriteString("B")}
       if Claim      == "1" { content.WriteString("C")}
       if Drawing     == "1" {content.WriteString("D")}
       if Amendement  =="1" {content.WriteString("X")}
       if Description  == "1" {content.WriteString("Y")}
       if Search_report  == "1" {content.WriteString("S")}
       /* +40*/
       usermd["content"] = content.String()
       data_type := string(st33[180:181])
       usermd["data_type"] = data_type
        
       // comp_meth := st33[181:183]

       // k_fac := st33[183:185]
       // Resolution := st33[185:187]
        
        s_fr_h:= st33[187:190] 
        s_fr_w:= st33[190:193] 
        nl_fr_h := st33[193:197]
        nl_fr_w :=st33[197:201]
        rotation_code := st33[201:202]
        // fr_x := st33[202:206]
        // fr_y := st33[206:210] 
        // fr_stat := st33[210:211]
        version := st33[211:214] 

        if string(version) == "V30" {
             buf1 = bytes.NewReader(buf[k+214:k+218])
             err = binary.Read(buf1,Little, &total_length) 
        } else {
           buf1 = bytes.NewReader(buf[k+214:k+218])
           err = binary.Read(buf1,Big, &total_length) 
        }
        //t := int(total_length)

        /* write tiff images */
        var img = new(bytes.Buffer)  
               
        _, err := io.WriteString(img, beHeader)            // magic number
         err = binary.Write(img, enc, uint32(ifdOffset))   // IFD offset
         err = binary.Write(img, enc, uint16(ifdLen))      // number of IFD entries
         
        err = binary.Write(img, enc, uint16(tImageWidth)) //  image Width
        err = binary.Write(img, enc, uint16(dtLong))      //  long
        err = binary.Write(img, enc, uint32(1))           //  value
        err = binary.Write(img, enc, getuint32(nl_fr_w))

        //Imagewidth,_ := strconv.Atoi(string(nl_fr_w))      
        //err = binary.Write(img, enc, uint32(Imagewidth))  // 

        err = binary.Write(img, enc, uint16(tImageLength)) // Image length
        err = binary.Write(img, enc, uint16(dtLong))       // long
        err = binary.Write(img, enc, uint32(1))            // value
        err = binary.Write(img, enc, getuint32(nl_fr_h))
       // ImageLength,_ := strconv.Atoi(string(nl_fr_h))      
       // err = binary.Write(img, enc, uint32(ImageLength))  // 
         
        err = binary.Write(img, enc, uint16(tCompression)) //  Compression
        err = binary.Write(img, enc, uint16(dtShort))      //  short 
        err = binary.Write(img, enc, uint32(1))            //  value
        err = binary.Write(img, enc, uint16(cG4))          //  CCITT Group 4 
        err = binary.Write(img, enc, uint16(0))            //  CCITT Group 4 
       
        err = binary.Write(img, enc, uint16(tPhotometricInterpretation)) //  Photometric
        err = binary.Write(img, enc, uint16(dtShort))        //  short 
        err = binary.Write(img, enc, uint32(1))              //  value
        err = binary.Write(img, enc, uint32(0))              //  white 

        err = binary.Write(img, enc, uint16(tStripOffsets))  //  StripOffsets 
        err = binary.Write(img, enc, uint16(dtLong))         //  long
        err = binary.Write(img, enc, uint32(1))              //  value
        err = binary.Write(img, enc, uint32(150))            //  0xA0

        err = binary.Write(img, enc, uint16(tOrientation))   // Orientation
        err = binary.Write(img, enc, uint16(dtShort))        //  short 
        err = binary.Write(img, enc, uint32(1)) 
        err = binary.Write(img, enc, getOrientation(rotation_code)) // rotation code
        err = binary.Write(img, enc, uint16(0)) 
      
        err = binary.Write(img, enc, uint16(tStripByteCounts)) //  StripbyteCounts
        err = binary.Write(img, enc, uint16(dtLong))          //  long
        err = binary.Write(img, enc, uint32(1))  
        err = binary.Write(img, enc, uint32(total_length))    //  image size
        
        err = binary.Write(img, enc, uint16(tXResolution))   // Xresolution
        err = binary.Write(img, enc, uint16(dtRational))     // rational
        err = binary.Write(img, enc, uint32(1))  
        err = binary.Write(img, enc, uint32(xoffset))          // 
        
        err = binary.Write(img, enc, uint16(tYResolution))     // Xresolution
        err = binary.Write(img, enc, uint16(dtRational))       //  Rational
        err = binary.Write(img, enc, uint32(1))  
        err = binary.Write(img, enc, uint32(yoffset))          // 

        err = binary.Write(img, enc, uint16(tResolutionUnit))  // Xresolution
        err = binary.Write(img, enc, uint16(dtShort))          //  value
        err = binary.Write(img, enc, uint32(1))  
        err = binary.Write(img, enc, uint16(3))               //  cm
        err = binary.Write(img, enc, uint16(0)) 

        err = binary.Write(img, enc, uint32(0))               // next IFD = 0

        // Xresolution value
        err = binary.Write(img, enc, getuint32(nl_fr_w)*10)
        err = binary.Write(img, enc, getuint32(s_fr_w)) 

        // Yresoluton value
        err = binary.Write(img, enc, getuint32(nl_fr_h)*10)
        err = binary.Write(img, enc, getuint32(s_fr_h)) 
        
        docid1        := strings.TrimLeft(string(docid)," ") 
        filename_o    := string(pub)+docid1+string(kc)+"_"+string(pagenum)
        usermdB,_     := json.Marshal(usermd)
        file_usermd_o := outputusermdDir + filename_o+".json" 
        file_tiff_o   := outputTiffDir + filename_o+".tiff"
    
        //write images
         for  r:=0; r< int(total_rec);r++ {
          l1 := ebc2asc.Ebc2asci(buf[k:k+5])
          long,_ = strconv.Atoi(string(l1))
          //buf1 = bytes.NewReader(buf[k+250:k+252])
          //err = binary.Read(buf1, Big, &imgl0)
          err = binary.Read(bytes.NewReader(buf[k+250:k+252]), Big, &imgl)
          //imgl := int(imgl0)
          img.Write(buf[k+252:k+252+int(imgl)])    
          k = k + long
          //fmt.Println(r,imgl, "img len=",img.Len())
        } 

        /* wrtite tiff to file */
        err = ioutil.WriteFile(file_tiff_o,img.Bytes(),0644)
        check(err)
        img.Reset() // 
        err = ioutil.WriteFile(file_usermd_o,usermdB,0644)
        check(err)
        
        l = k   
       }
       
    }
   buf = buf[:0]
   fmt.Println(len(buf))
 }  // for filename
}
