(defproject prowler "0.1.0-SNAPSHOT"
  :description "CI status tracker"
  :url "https://github.com/de1ux/prowler"
  :main prowler.core
  :license {:name "Eclipse Public License"
            :url "http://www.eclipse.org/legal/epl-v10.html"}
  :plugins [[lein-cljfmt "0.5.3"]]
  :dependencies [[org.clojure/clojure "1.8.0"]
                 [clj-http "2.1.0"]
                 [org.clojure/data.json "0.2.6"]
                 [clj-time "0.11.0"]
                 [environ "1.0.2"]
                 [cheshire "5.6.3"]])
