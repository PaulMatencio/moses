package sindexd

import (
	"bytes"
	"encoding/json"
	// goLog "github.com/moses/user/goLog"
	goLog "github.com/s3/gLog"
	"net/http"
)

type Stats struct {
	Lowlevel  int `json:"lowlevel,omitempty"`
	Highlevel int `json:"highlevel,omitempty"`
	Cache     int `json:"cache,omitempty"`
	Reset     int `json:"reset,omitempty"`
}

type Get_Stats struct {
	Stats `json:"stats"`
}

func (s *Get_Stats) GetStats(client *http.Client) (*http.Response, error) {
	/*
	   [{ "hello":{ "protocol": "sindexd-1"} }.
	   { "stats": {"lowlevel": 1,"highlevel": 1,"cache": 1,"reset": 0 }} ]
	*/
	if sj, err := json.Marshal(s); err == nil {
		myreq := [][]byte{[]byte(AG), []byte(HELLO), []byte(V), sj, []byte(AD)}
		request := bytes.Join(myreq, []byte(""))
		return PostRequest(client, request)
	} else {
		return nil, err
	}
}

func PrintStats(f string, resp *http.Response) {
	//gstats := new(GetStats)
	//json.Unmarshal(GetBody(resp), &gstats)
	goLog.Info.Println(f, string(GetBody(resp)))
	//goLog.Info.Println(f, gstats.Stats.Highlevel)

}

type GetStats struct {
	Status   int    `json:"status"`
	Protocol string `json:"protocol"`
	IndexID  string `json:"index_id"`
	Stats    struct {
		UptimeS int `json:"uptime_s"`
		Rusage  struct {
			UtimeMs  int `json:"utime_ms"`
			StimeMs  int `json:"stime_ms"`
			Maxrss   int `json:"maxrss"`
			Ixrss    int `json:"ixrss"`
			Idrss    int `json:"idrss"`
			Isrss    int `json:"isrss"`
			Minflt   int `json:"minflt"`
			Majflt   int `json:"majflt"`
			Nswap    int `json:"nswap"`
			Inblock  int `json:"inblock"`
			Oublock  int `json:"oublock"`
			Msgsnd   int `json:"msgsnd"`
			Msgrcv   int `json:"msgrcv"`
			Nsignals int `json:"nsignals"`
			Nvcsw    int `json:"nvcsw"`
			Nivcsw   int `json:"nivcsw"`
		} `json:"rusage"`
		Highlevel struct {
			HelloOps     int `json:"hello_ops"`
			HelloAvg     int `json:"hello_avg"`
			CreateOps    int `json:"create_ops"`
			CreateAvg    int `json:"create_avg"`
			LoadOps      int `json:"load_ops"`
			LoadAvg      int `json:"load_avg"`
			AddOps       int `json:"add_ops"`
			AddAvg       int `json:"add_avg"`
			GetOps       int `json:"get_ops"`
			GetAvg       int `json:"get_avg"`
			GetPrefixOps int `json:"get-prefix_ops"`
			GetPrefixAvg int `json:"get-prefix_avg"`
			DeleteOps    int `json:"delete_ops"`
			DeleteAvg    int `json:"delete_avg"`
			DropOps      int `json:"drop_ops"`
			DropAvg      int `json:"drop_avg"`
			StatsOps     int `json:"stats_ops"`
			StatsAvg     int `json:"stats_avg"`
			ConfigOps    int `json:"config_ops"`
			ConfigAvg    int `json:"config_avg"`
			CopyOps      int `json:"copy_ops"`
			CopyAvg      int `json:"copy_avg"`
			RenameOps    int `json:"rename_ops"`
			RenameAvg    int `json:"rename_avg"`
			DrainOps     int `json:"drain_ops"`
			DrainAvg     int `json:"drain_avg"`
			CommitOps    int `json:"commit_ops"`
			CommitAvg    int `json:"commit_avg"`
			RollbackOps  int `json:"rollback_ops"`
			RollbackAvg  int `json:"rollback_avg"`
			OtherOps     int `json:"other_ops"`
			OtherAvg     int `json:"other_avg"`
		} `json:"highlevel"`
		CacheStructure struct {
			MaxSize         int64 `json:"max_size"`
			Size            int   `json:"size"`
			RecomputedSize  int   `json:"recomputed_size"`
			Count           int   `json:"count"`
			RecomputedCount int   `json:"recomputed_count"`
			BusyCount       int   `json:"busy_count"`
		} `json:"cache_structure"`
		Cache struct {
			Timestamp string `json:"timestamp"`
			Stats     struct {
				Open struct {
					Count struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"count"`
					Concurrent struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"concurrent"`
					MaxConcurrent struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"max_concurrent"`
					Errors struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"errors"`
					AvgMs struct {
						Value       float64 `json:"value"`
						Type        string  `json:"type"`
						Unit        string  `json:"unit"`
						Description string  `json:"description"`
						Dependances struct {
							TimeSpent string `json:"time_spent"`
							Count     string `json:"count"`
						} `json:"dependances"`
					} `json:"avg_ms"`
					SlowestMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"slowest_ms"`
					TotalMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"total_ms"`
				} `json:"open"`
				Update struct {
					Count struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"count"`
					Concurrent struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"concurrent"`
					MaxConcurrent struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"max_concurrent"`
					Errors struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"errors"`
					AvgMs struct {
						Value       float64 `json:"value"`
						Type        string  `json:"type"`
						Unit        string  `json:"unit"`
						Description string  `json:"description"`
						Dependances struct {
							TimeSpent string `json:"time_spent"`
							Count     string `json:"count"`
						} `json:"dependances"`
					} `json:"avg_ms"`
					SlowestMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"slowest_ms"`
					TotalMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"total_ms"`
				} `json:"update"`
				Delete struct {
					Count struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"count"`
					Concurrent struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"concurrent"`
					MaxConcurrent struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"max_concurrent"`
					Errors struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"errors"`
					AvgMs struct {
						Value       float64 `json:"value"`
						Type        string  `json:"type"`
						Unit        string  `json:"unit"`
						Description string  `json:"description"`
						Dependances struct {
							TimeSpent string `json:"time_spent"`
							Count     string `json:"count"`
						} `json:"dependances"`
					} `json:"avg_ms"`
					SlowestMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"slowest_ms"`
					TotalMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"total_ms"`
				} `json:"delete"`
				LlCheck struct {
					Count struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"count"`
					Concurrent struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"concurrent"`
					MaxConcurrent struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"max_concurrent"`
					Errors struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"errors"`
					AvgMs struct {
						Value       float64 `json:"value"`
						Type        string  `json:"type"`
						Unit        string  `json:"unit"`
						Description string  `json:"description"`
						Dependances struct {
							TimeSpent string `json:"time_spent"`
							Count     string `json:"count"`
						} `json:"dependances"`
					} `json:"avg_ms"`
					SlowestMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"slowest_ms"`
					TotalMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"total_ms"`
				} `json:"ll_check"`
				LlUploadMd struct {
					Count struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"count"`
					Concurrent struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"concurrent"`
					MaxConcurrent struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"max_concurrent"`
					Errors struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"errors"`
					AvgMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
						Dependances struct {
							TimeSpent string `json:"time_spent"`
							Count     string `json:"count"`
						} `json:"dependances"`
					} `json:"avg_ms"`
					SlowestMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"slowest_ms"`
					TotalMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"total_ms"`
				} `json:"ll_upload_md"`
				LlDelete struct {
					Count struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"count"`
					Concurrent struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"concurrent"`
					MaxConcurrent struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"max_concurrent"`
					Errors struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"errors"`
					AvgMs struct {
						Value       float64 `json:"value"`
						Type        string  `json:"type"`
						Unit        string  `json:"unit"`
						Description string  `json:"description"`
						Dependances struct {
							TimeSpent string `json:"time_spent"`
							Count     string `json:"count"`
						} `json:"dependances"`
					} `json:"avg_ms"`
					SlowestMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"slowest_ms"`
					TotalMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"total_ms"`
				} `json:"ll_delete"`
				LlDownload struct {
					Count struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"count"`
					Concurrent struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"concurrent"`
					MaxConcurrent struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"max_concurrent"`
					Errors struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"errors"`
					AvgMs struct {
						Value       float64 `json:"value"`
						Type        string  `json:"type"`
						Unit        string  `json:"unit"`
						Description string  `json:"description"`
						Dependances struct {
							TimeSpent string `json:"time_spent"`
							Count     string `json:"count"`
						} `json:"dependances"`
					} `json:"avg_ms"`
					SlowestMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"slowest_ms"`
					TotalMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"total_ms"`
					TotalSize struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"total_size"`
					SlicesBySize []struct {
						LessBytesThan int `json:"less_bytes_than"`
						Count         struct {
							Value       int    `json:"value"`
							Type        string `json:"type"`
							Unit        string `json:"unit"`
							Description string `json:"description"`
						} `json:"count"`
						Errors struct {
							Value       int    `json:"value"`
							Type        string `json:"type"`
							Unit        string `json:"unit"`
							Description string `json:"description"`
						} `json:"errors"`
						AvgMs struct {
							Value       int    `json:"value"`
							Type        string `json:"type"`
							Unit        string `json:"unit"`
							Description string `json:"description"`
							Dependances struct {
								TimeSpent string `json:"time_spent"`
								Count     string `json:"count"`
							} `json:"dependances"`
						} `json:"avg_ms"`
						SlowestMs struct {
							Value       int    `json:"value"`
							Type        string `json:"type"`
							Unit        string `json:"unit"`
							Description string `json:"description"`
						} `json:"slowest_ms"`
						TotalMs struct {
							Value       int    `json:"value"`
							Type        string `json:"type"`
							Unit        string `json:"unit"`
							Description string `json:"description"`
						} `json:"total_ms"`
						TotalSize struct {
							Value       int    `json:"value"`
							Type        string `json:"type"`
							Unit        string `json:"unit"`
							Description string `json:"description"`
						} `json:"total_size"`
					} `json:"slices_by_size"`
				} `json:"ll_download"`
				LlUpload struct {
					Count struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"count"`
					Concurrent struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"concurrent"`
					MaxConcurrent struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"max_concurrent"`
					Errors struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"errors"`
					AvgMs struct {
						Value       float64 `json:"value"`
						Type        string  `json:"type"`
						Unit        string  `json:"unit"`
						Description string  `json:"description"`
						Dependances struct {
							TimeSpent string `json:"time_spent"`
							Count     string `json:"count"`
						} `json:"dependances"`
					} `json:"avg_ms"`
					SlowestMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"slowest_ms"`
					TotalMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"total_ms"`
					TotalSize struct {
						Value       int64  `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"total_size"`
					SlicesBySize []struct {
						LessBytesThan int `json:"less_bytes_than"`
						Count         struct {
							Value       int    `json:"value"`
							Type        string `json:"type"`
							Unit        string `json:"unit"`
							Description string `json:"description"`
						} `json:"count"`
						Errors struct {
							Value       int    `json:"value"`
							Type        string `json:"type"`
							Unit        string `json:"unit"`
							Description string `json:"description"`
						} `json:"errors"`
						AvgMs struct {
							Value       int    `json:"value"`
							Type        string `json:"type"`
							Unit        string `json:"unit"`
							Description string `json:"description"`
							Dependances struct {
								TimeSpent string `json:"time_spent"`
								Count     string `json:"count"`
							} `json:"dependances"`
						} `json:"avg_ms"`
						SlowestMs struct {
							Value       int    `json:"value"`
							Type        string `json:"type"`
							Unit        string `json:"unit"`
							Description string `json:"description"`
						} `json:"slowest_ms"`
						TotalMs struct {
							Value       int    `json:"value"`
							Type        string `json:"type"`
							Unit        string `json:"unit"`
							Description string `json:"description"`
						} `json:"total_ms"`
						TotalSize struct {
							Value       int    `json:"value"`
							Type        string `json:"type"`
							Unit        string `json:"unit"`
							Description string `json:"description"`
						} `json:"total_size"`
					} `json:"slices_by_size"`
				} `json:"ll_upload"`
				WbWrite struct {
					Count struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"count"`
					Concurrent struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"concurrent"`
					MaxConcurrent struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"max_concurrent"`
					Errors struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"errors"`
					AvgMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
						Dependances struct {
							TimeSpent string `json:"time_spent"`
							Count     string `json:"count"`
						} `json:"dependances"`
					} `json:"avg_ms"`
					SlowestMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"slowest_ms"`
					TotalMs struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"total_ms"`
					TotalSize struct {
						Value       int    `json:"value"`
						Type        string `json:"type"`
						Unit        string `json:"unit"`
						Description string `json:"description"`
					} `json:"total_size"`
					SlicesBySize []struct {
						LessBytesThan int `json:"less_bytes_than"`
						Count         struct {
							Value       int    `json:"value"`
							Type        string `json:"type"`
							Unit        string `json:"unit"`
							Description string `json:"description"`
						} `json:"count"`
						Errors struct {
							Value       int    `json:"value"`
							Type        string `json:"type"`
							Unit        string `json:"unit"`
							Description string `json:"description"`
						} `json:"errors"`
						AvgMs struct {
							Value       int    `json:"value"`
							Type        string `json:"type"`
							Unit        string `json:"unit"`
							Description string `json:"description"`
							Dependances struct {
								TimeSpent string `json:"time_spent"`
								Count     string `json:"count"`
							} `json:"dependances"`
						} `json:"avg_ms"`
						SlowestMs struct {
							Value       int    `json:"value"`
							Type        string `json:"type"`
							Unit        string `json:"unit"`
							Description string `json:"description"`
						} `json:"slowest_ms"`
						TotalMs struct {
							Value       int    `json:"value"`
							Type        string `json:"type"`
							Unit        string `json:"unit"`
							Description string `json:"description"`
						} `json:"total_ms"`
						TotalSize struct {
							Value       int    `json:"value"`
							Type        string `json:"type"`
							Unit        string `json:"unit"`
							Description string `json:"description"`
						} `json:"total_size"`
					} `json:"slices_by_size"`
				} `json:"wb_write"`
			} `json:"stats"`
			TotalRead struct {
				Value       int    `json:"value"`
				Type        string `json:"type"`
				Unit        string `json:"unit"`
				Description string `json:"description"`
			} `json:"total_read"`
			TotalWritten struct {
				Value       int    `json:"value"`
				Type        string `json:"type"`
				Unit        string `json:"unit"`
				Description string `json:"description"`
			} `json:"total_written"`
		} `json:"cache"`
	} `json:"stats"`
}
