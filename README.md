Prowler
=======

### Setup
1) Go to System Preferences > Security & Privacy, allow apps downloaded from anywhere

2) Go to https://github.com/settings/tokens and generate a new personal access token. Allow access to everything except deleting repos. Save this token somewhere for the next step

3) Save .prowler.conf to your home directory. The configuration file is an ordered array with three expected values:
    a) your username
    b) the repos you wish to track (prowl)
    c) your access token

4) Edit .prowler.conf to include your name, access token, and desired repos. Note that the repos must be entered with escaping backslashes (see example)

5) Save and exit

6) Unzip Prowler.zip

7) Move Prowler.app to your Applications

8) Start Prowler

### Usage
Click on the title to go to the PR.

Click on the CI labels to go to the CI.

Labels marked in YELLOW are currently running CIs.

Labels marked in RED are failed CI runs.

Labels marked in GREEN are passing CI runs.

Titles marked with a red circle and bar mean they have merge conflicts.
