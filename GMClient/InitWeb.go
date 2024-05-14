package main

import (
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"git.code4.in/mobilegameserver/config"
)

var (
	tplMap map[string]*template.Template = make(map[string]*template.Template)
)

type HttpHandlerFunc func(w http.ResponseWriter, r *http.Request)

func InitWebApp() bool {
	static, _ := filepath.Abs(config.GetConfigStr("static"))
	tplPath, _ := filepath.Abs(config.GetConfigStr("views"))

	http.Handle("/js/", http.FileServer(http.Dir(static)))
	http.Handle("/css/", http.FileServer(http.Dir(static)))
	http.Handle("/img/", http.FileServer(http.Dir(static)))
	http.Handle("/font/", http.FileServer(http.Dir(static)))
	http.Handle("/form/", http.FileServer(http.Dir(static)))
	http.Handle("/My97DatePicker/", http.FileServer(http.Dir(static)))
	http.Handle("/assets/", http.FileServer(http.Dir(static)))

	InitStaticViews(tplPath)
	http.HandleFunc("/logout", HandleLogout)
	http.HandleFunc("/", HandleRoot)
	http.HandleFunc("/download", HandleDownloadFileHttp)
	return true
}

func InitStaticViews(views string) {
	if views[len(views)-1] != '/' {
		views += "/"
	}
	filepath.Walk(views, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return err
		} else if !info.IsDir() {
			basename := "/" + filepath.Base(path)
			t, err := template.ParseFiles(path)
			if err != nil {
				return err
			}
			tplMap[basename] = t
			if basename == "/login.html" {
				http.HandleFunc(basename, HandleLogin)
			} else {
				http.HandleFunc(basename, HandleOther)
			}
		}
		return nil
	})
	http.HandleFunc("/favicon.ico", HandlerIcon)
}

func AddHandlerWithTpl(tplPath, url string, hfunc http.HandlerFunc) {
	t, err := template.ParseFiles(tplPath + url + ".html")
	if err != nil {
		return
	}
	tplMap[url] = t
	http.HandleFunc(url, hfunc)
}

func HandlerIcon(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/img/"+r.URL.Path, http.StatusFound)
}

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login.html", http.StatusFound)
}

func HandleOther(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	task := GetGMClientTask(r, 0)
	basename := filepath.Base(r.Referer())
	if task == nil || (r.URL.Path != "/home.html" && basename != "home.html") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tp := task.Data.GetPri()
	if tp == 0 || task.Data.GetWorkstate() == 1 {
		tp = 1
	} else {
		tp = tp & 2
	}
	if config.GetConfigStr("debug") != "false" {
		views, _ := filepath.Abs(config.GetConfigStr("views"))
		if views[len(views)-1] != '/' {
			views += "/"
		}
		t, err := template.ParseFiles(views + r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		t.Execute(w, map[string]interface{}{"Username": task.Name, "T": tp, "Pri": task.Data.GetPri(), "GameList": task.GameList})
	} else {
		tplMap[r.URL.Path].Execute(w, map[string]interface{}{"Username": task.Name, "GameList": task.GameList, "T": tp, "Pri": task.Data.GetPri()})
	}
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	task := GetGMClientTask(r, 0)
	if task != nil {
		http.Redirect(w, r, "/home.html", http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	xsrfStr := XSRFFormHTML(w, config.GetConfigStr("secret_key"))
	tplMap["/login.html"].Execute(w, map[string]string{"_xsrf": xsrfStr})
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	task := GetGMClientTask(r, 0)
	if task == nil {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	ClearCookie(w, r, "id")
	task.Error("用户主动下线")
	task.RemoveMe()
	task.RequestClose("用户主动下线")
	http.Redirect(w, r, "/login.html", http.StatusFound)
}
