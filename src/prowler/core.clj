(ns prowler.core
  (:gen-class)
  (:require [clj-http.client :as http])
  (:require [clojure.data.json :as json])
  (:require [clojure.java.io :as io])
  (:require [clj-time.format :as f])
  (:require [environ.core :refer [env]])
  (:require [cheshire.core :refer :all])
  (:use [clojure.string :as s :exclude [reverse replace]]))

; Constants that should go somewhere else
(def DEFAULT_TOKEN "a github api key")
(def DEFAULT_REPOS ["de1ux/prowler" "you/coolRepo"])
(def DEFAULT_USERNAME "your github user name")
(def GITHUB_BASE_URL "https://api.github.com")
(def CONFIG_NAME ".prowler.conf")

; Every CI reports a different state keyword, so try to corral them into
; success/failure/pending categories
(def SUCCESS_STATES ["success"])
(def FAILURE_STATES ["failure" "failed" "error"])
(def PENDING_STATES ["pending"])

; getRequest returns a json-decoded result from a GET request to url
(defn getRequest [url] (json/read-str
                        (:body
                         (http/get url))))

; fetchRepo performs a getRequest on a repo given a repo name and token
(defn fetchRepo [repo token] (getRequest (format "%s/repos/%s/pulls?oauth_token=%s" GITHUB_BASE_URL repo token)))

; url is the raw url parsed from the github api
(defn fetchPr [url token] (getRequest (format "%s?oauth_token=%s" url token)))

; filterPullRequestsByUser filters pull requests by user
(defn filterPullRequestsByUser [repoJson, userName showAllPrs] (keep #(if (or (= userName (get-in % ["user" "login"])) showAllPrs) %) repoJson))

(defn getLatestStatus [statusJson service] (last (sort-by
                                                  #(f/parse (% "updated_at"))
                                                  (filter #(s/includes? (get-in % ["target_url"]) service) (remove #(nil? (get-in % ["target_url"])) statusJson)))))

(defn parseStatus [statusJson services] (map (fn [service] (getLatestStatus statusJson service)) services))

(defn fetchStatus [pullRequest repo token services] {:title (pullRequest "title")
                                                     :url (pullRequest "html_url")
                                                     :ci (parseStatus (getRequest (format "%s/repos/%s/statuses/%s?oauth_token=%s" GITHUB_BASE_URL
                                                                                          repo (get-in pullRequest ["head" "sha"]) token)) services)
                                                     :mergeable ((fetchPr (pullRequest "url") token) "mergeable")})

(defn fetchStatuses [pullRequests repo token services] (map (fn [pullRequest] (fetchStatus pullRequest repo token services)) pullRequests))

(defn fetchRecents [userName token] (getRequest (format "%s/users/%s/events?oauth_token=%s" GITHUB_BASE_URL userName token)))

(defn getColorFromStatus [status successStates pendingStates failureStates] (if (some #{status} successStates) " color=#00bb00" (if (some #{status} pendingStates) " color=#f89406" " color=#bb0000")))

(defn statusToBitBarFormat [ci successStates pendingStates failureStates] (str " | href=" (get-in ci ["target_url"]) (getColorFromStatus (get-in ci ["state"]) successStates pendingStates failureStates)))

(defn prToStatusesToBitBarFormat [pr successStates pendingStates failureStates] (s/join ";" (map (fn [ci] (str "⌊ " (get-in ci ["context"]) (statusToBitBarFormat ci successStates pendingStates failureStates))) (remove #(nil? (get-in % ["context"])) (get-in pr [:ci])))))

(defn prsToBitBarFormat [repo successStates pendingStates failureStates hideMergeConflicts] (s/join ";" (map (fn [pr] (str (get-in pr [:title]) " " (if (or (get-in pr [:mergeable]) hideMergeConflicts) "" "🚫") " | href=" (get-in pr [:url]) ";" (prToStatusesToBitBarFormat pr successStates pendingStates failureStates))) (get-in repo [:prs]))))

(defn toBitBarFormat [manifest successStates pendingStates failureStates hideMergeConflicts] (str "❦;" (s/join ";" (map (fn [repo] (str "---;" (get-in repo [:repo]) " | size=20" ";" (prsToBitBarFormat repo successStates pendingStates failureStates hideMergeConflicts))) manifest))))

(defn createManifest [userName repos token services successStates pendingStates failureStates hideMergeConflicts showAllPrs] (println (toBitBarFormat (map (fn [repo]
                                                                                                                               {:repo repo
                                                                                                                                :prs (fetchStatuses (filterPullRequestsByUser (fetchRepo repo token) userName showAllPrs) repo token services)
                                                                                                                                :recents (fetchRecents userName token)}) repos) successStates pendingStates failureStates hideMergeConflicts)))

(defn getConfigDefaults [] {:username DEFAULT_USERNAME :repos DEFAULT_REPOS :token DEFAULT_TOKEN :services [] :successStates SUCCESS_STATES :pendingStates PENDING_STATES :failureStates FAILURE_STATES :hideMergeConflicts false :showAllPrs false})

(defn getConfigPath [] (str (System/getenv "HOME") "/" CONFIG_NAME))

(defn loadConfig [& args] (json/read-str (slurp (getConfigPath))))

(defn createConfig [] (spit (getConfigPath) (generate-string (getConfigDefaults) {:pretty true})))

(defn getConfig [] (loadConfig (if (not (.exists (io/as-file (getConfigPath)))) (createConfig))))

(defn explodeConfig [config] [(get-in config ["username"]) (get-in config ["repos"]) (get-in config ["token"]) (get-in config ["services"]) (get-in config ["successStates"]) (get-in config ["pendingStates"]) (get-in config ["failureStates"]) (get-in config ["hideMergeConflicts"]) (get-in config ["showAllPrs"])])

(defn -main [] (apply createManifest (explodeConfig (getConfig))))
