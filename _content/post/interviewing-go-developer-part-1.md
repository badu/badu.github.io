---
title: Interview Questions for Go Developer Position
tags: ["Go", "Developer", "Interview"]
date: 2018-11-18
description: Measuring And Classifying Go Developer Knowledge
---

#### Foreword

There is no such thing as essential interview questions. Despite the fact that I'm trying to help you, formulate questions and/or answers, I cannot help you point in the right direction, since each job has specific requirements and each person is unique.

My word of advising is don't jump to conclusions, just because someone didn't answer your question like you expected you to do. And there are people out there - myself included - who are able to explain you a concept without remembering the actual name or the acronym.

Also, do not make the mistake of comparing Go with other languages, unless you want to shot yourself in the foot. Asking questions like "Explain how arrays in Go works differently than C?" would imply that both of you know C and Go, which is not always the case.
Expecting an answer like "arrays are values, assigning one array to another copies all the elements" means that your evaluation would have a reasonable chance to get the wrong results (just like you are expecting the candidate to fail). Ask the smart questions, like "if I pass an array to a function, can I modify that array? After the function returns, would that array reflect the changes made inside that function?".

#### Scale Points Rating

Your technical interview should be based on a 0, 1, 3 rating points for each question. Zero points would go for unable to answer (nor "how", nor "why"), one point for providing a basic answer ("how") and three points for having deep knowledge ("why"). There are questions that should go on a different scoring table (marked Optional Question or OQ), just in case you need to choose the best of the best candidate in case of ties.

#### Questions

If you Google it, you'll find a lot of websites that provide material for both study and questioning candidates. It is, of course, up to you to make a strategy and prepare for an interview.

To give you an example, if someone would ask me "What is Go?" during an interview, I would be tempted to answer "C on steroids" while doubting that the interviewer is well prepared for the job of evaluating me.

On the other hand, if the questions are about particular programming doctrines and less about Go, one would ask itself if the job position is really about a Go development position since the measurements are made by asking something else. To give you an example, I've been asked about Domain Driven Design and my answer was something like "never heard of it" while trying to provide a valid answer on what I think it might be. Penalty points do not apply if you measure Go skills and know-how while control checks on doctrines and principles.

In the end, the difference between an experienced developer and a young padawan is made by curiosity and occasional need to go deeper down the rabbit whole. One might know how to use maps and slices in Go, while never had the curiosity of finding out how maps internally work.

Again, depends what you measure.

#### General Questions

Q: How do you swap two variables?

A: `a , b = b , a` [1 point]

---

Q: The above operation can have side effects?

A: No [1 point]

A: No. Go is storing those in temporary variables, so you can swap multiple values `a,b,c = c,a,b` [3 points]

---

OQ: Can you swap two numbers without using a temporary variable?

A: `a = a ^ b; b = b ^ a; a = a ^ b` [3 points]

---

Q: Is `for i := range arr{}` correct?

A: Yes [1 point]

---

Q: What is the difference between array and slice in Go?

A: Arrays are fixed-length [1 point]

A: Arrays are fixed-length and slices are reference types, arrays are values [3 points]

---

Q: Is it possible to change a string once created?

A: No, you get a copy [1 point]

A: No, they are immutable [3 points]

---

Q: What is the purpose of the `fallthrough` statement?

A: Transfers control to the next case [1 point]

---

Q: Do you know a base package that uses `fallthrough` statement?

A: [Yes](https://golang.org/src/encoding/ascii85/ascii85.go#L43) [1 point]

Note : candidate can provide something similar.

---

Q: Can we return multiple values from a function?

A: Yes [1 point]

---

Q: What are the main characteristics of a function closures in Go?

A: It is anonymous and have access to variables from outer scope [1 point]

Note : regarding outer scope, candidate can answer "references variables outside its body"

---

Q: Is there a limit of execution for recursive function call?

A: Yes, see [this](https://play.golang.org/p/XzZsmGLgIp) [1 point]

---

Q: Can we optimize recursive function calls?

A: Yes, using channels [1 point]

A: Yes, using channels and candidate also mentions "tail call" in his answer.
Example. [here](https://play.golang.org/p/e3f0SjPXhVc) [3 point]

---

#### Slices/Arrays Questions

Q : Having `var a []int` and `a := []int{}`, which is more preferable?

A : `var a []int` [1 point]

A : `var a []int` because it doesn't allocate memory [3 points]

---

Q: If I pass an array to a function, can I modify it inside that function?

A: No [1 point]

A: No, you get a copy of that array [3 points]

---

Q: If I pass an slice to a function, can I modify it inside that function?

A: Yes [1 point]

A: Yes, because you don't get a copy of it because they are reference types [3 points]

---

Q: How do you copy a slice?

A: Using copy() [1 point]

---

Q: What is `x, a = a[len(a)-1], a[:len(a)-1]` doing?

A: Pop from a slice [1 point]

---

Q: What is `a = append([]int{x}, a...)` doing?

A: Push front (also called unshift) [1 point]

---

Q: How do you delete an element from a slice?

A: `a = append(a[:i], a[i+1:]...)` or `a = a[:i+copy(a[i:], a[i+1:])]` [1 point]

A: `copy(a[i:], a[i+1:]); a[len(a)-1] = nil; a = a[:len(a)-1]` [3 points]

Note : If the type of the element is a pointer or a struct with pointer fields, which need to be garbage collected, the above implementation of Delete have a potential memory leak problem, for this reason the second answer has 3 points.

---

Q: How would you implement a stack and a queue in Go?

A: [here](https://play.golang.org/p/vRighjOw0iT) [1 point each]

---

Q: Why one would use or not use `container/list` ?

A: Slower performance, compared to slices iteration pattern [1 point]

A: Slower performance because “Always use a slice.” said Dave Cheney [3 points]

A: Another possibility to implement a queue is to use buffered channels, but this is never a good idea [3 points]

Note : "Novices are sometimes tempted to use buffered channels within a single goroutine as a queue, lured by their pleasingly simple syntax, but this is a mistake. Channels are deeply connected to goroutine scheduling, and without another goroutine receiving from the channel, a sender—and perhaps the whole program—risks becoming blocked forever. If all you need is a simple queue, make one using a slice"

---

#### Maps Questions

Q: Do maps have a fixed order for their keys?

A: No [1 point]

---

Q: How we can display a fixed order of map keys?

A: Put the keys in a slice and sort it [1 point]

---

Q: How do you copy a map?

A: By traversing its keys [1 point]

---

Q: What are the uses of empty struct{}?

A: Save memory (`a:=struct{}{}; println(unsafe.Sizeof(a))` which is zero) [1 point]

---

Q: Provide examples of the empty struct{} usage

A: Implementing dataset or seen hash [1 point]
```
seen := make(map[string]struct{})
for _, ok := seen[v]; ok {
    // set the visited "flag"
    seen[v] = struct{}{}
}
```

A: Implementing dataset, seen hash as above, or grouping methods and no intermediary data, need a channel to signal an event, but do not really need to send any data. [3 points]

```
// grouping
type Doer struct{}
func (d Doer) DoSomething(){
    println("Done")
}
func (d Doer) DoSomethingElse(){
    println("Something Else Done")
}
// or signaling
func worker(ch chan struct{}) {
	// Receive a message from the main program.
	<-ch
	println("roger-roger")
	// Send a message to the main program.
	close(ch)
}
```
---

Q: How we can delete an entry from a map?

A: using delete() function [1 point]

---

#### Optional (Advanced) questions

OQ: Select is random or sequential ?

A: When multiple cases in a select statement are ready, one of them will be [executed at random](https://play.golang.org/p/vJ6VhVl9YY). [3 points]

---

OQ: Go local variables are allocated on the stack or heap?

A: `escape analysis` : compiler will automatically decide whether to put a variable on the stack or put it on the heap [3 points]


---

Note for advanced discussions : Solutions for Elements of Programming Interviews problems written in Go can be found [here](https://github.com/mrekucci/epi).

---

To be continued.
