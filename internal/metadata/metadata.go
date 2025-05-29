package metadata

//func GetMetadata(path string) (*models.Archive, error) {
//	if fsutil.Is7z(path) {
//		zip, err := sevenzip.OpenReader(path)
//		if err != nil {
//			log.Println(err)
//		}
//		defer zip.Close()
//
//		for _, file := range zip.File {
//			switch filepath.Base(file.Name) {
//			case "ComicInfo.xml":
//				content, err := fsutil.Read7z(*file)
//				if err != nil {
//					return nil, err
//				}
//
//				data, err := ComicInfoUnmarshal(content)
//				if err != nil {
//					return nil, err
//				}
//
//				return data, nil
//			}
//		}
//	} else {
//		zip, err := zip.OpenReader(path)
//		if err != nil {
//			log.Println(err)
//		}
//		defer zip.Close()
//
//		for _, file := range zip.File {
//			switch filepath.Base(file.Name) {
//			case "ComicInfo.xml":
//				content, err := fsutil.ReadZip(*file)
//				if err != nil {
//					return nil, err
//				}
//
//				data, err := ComicInfoUnmarshal(content)
//				if err != nil {
//					return nil, err
//				}
//
//				return data, nil
//			}
//		}
//	}
//	return nil, nil
//}
