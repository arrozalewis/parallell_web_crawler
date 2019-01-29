1. Usage and I/O
installation:
https://golang.org/doc/install

running the program:
> time go run monster.go [input file] [output file]

cmd line arguemnt io files
(input file) should be a .txt file in working dir 
             format of file contains one url per line w/o "" marks
             urls end in new line character and have no preceeding/trailing spaces
             ex:  
             BOF 
             https://www.indeed.com/viewjob?jk=a741ef8b463aff95&from=recjobs&vjtk=1d1spi9ggbs56800
             https://www.indeed.com/viewjob?jk=0d245a4665d330e2&from=recjobs&vjtk=1d1spi9ggbs56800
             https://www.indeed.com/viewjob?jk=830414c7fcd79650&from=recjobs&vjtk=1d0vtotigbr0h800
             EOF

(output file)should be name for .txt file which we will write the output of our json repsonse 

> time go run monster.go in.txt out.txt 
should create or overwrite out.txt with json repsonse and display program
runtime & lines read after finish

2. Discussion of technologies used
Service implements a REST API which services GET & POST Requests
Made specific use of imports from the go net/http package & encoding/json
to handle json requests and responses 

3. A section demonstrating that you met the requirements
completion time for (file) length of 100
(on 4 cores)
real	0m4.414s
user	0m0.763s
sys	0m0.276s

(one single core)
real	0m14.308s
user	0m0.887s
sys	0m0.337s`


Efficiency = Speedup / cores = (14.308s/4.41s)/4 = 3.244

system/user/real time usage
One core vs 4 core performance on macbook pro

4. For algorithms and datastructures, time and space complexity
Algorithms
depth first search for traversing webpage DOM elements 
O(|vertex| + |edges|) 

Requests are processed linearly giving overall program complexity of 
O(urls(|vertex| + |edges|) or approx O(url(|vertex|)) due to sparse graph 

Data Structures
Channels & WaitGroups for each request were executed in parallel 
  WaitGroups used to sync incoming repsonses from url requests  
  For simplicity I passed two channels both uni-directional 
  one in<- and one <-out to handle flow and data from string to json structs
  so each core can crawl one or more urls

Channels also used to exploit concurrency when traversing the html elements
  Title, Loc, Company, Url pulled from DOM and communicated bi-directionally
  w/ single change for specific fields (company was in seperate HTML tag)
  Select statements in loop used to sync channels that are ready to send 

5. Other Considerations
Ensure that monster program is consuming majority of CPU% to ensure program
executes optimally. 4 core processor should consume well over 350% of CPU.
While running other apps I found my program to slow down significantly
sometimes varying as much to 15s to 30s to complete.

-Lewis J 1/25/19
