---
title: Comparing Go with Other Languages is Wrong
tags: ["Golang", "Programming", "Languages"]
date: 2017-11-15
description: Java is to Javascript like Car is to Carpet
---

## TL;DR

"Java is to Javascript like Car is to Carpet." - Collective Genius

## My Computer Language is Better than Yours

Some programming languages were - let's say - discovered over time, whereas others were created academically by carefully selecting the features that would provide solutions to a set of problems. There’s nothing terribly new about spawning programming languages. However, haters gonna hate and will always find one, in an office near you or over the Internet.

I'm not talking here about ranking languages. Comparing languages - that's this note about.

In my morning routine, I'm reading a lot of articles on various subjects, mostly related with Go and Java : once a week, I get across an article that compares Go to Scala, Java, you name it.

Comparing programming languages is like going to kindergarten all over again. Kids start arguing over the language features without having a clue what led the authors of that programming language to make an option over another option. Complaining that Go has no generics is from that mindset, where a young mind doesn't understand that Go is not object oriented, therefore there is no need for generics.

Go was an explicitly engineered language intended to solve problems with existing languages and tools while natively taking advantage of modern hardware architectures. It has been designed not only with teams of developers in mind, but also long-term maintainability. Comparing it to Java (which I know well) is wrong for the simple reason that they are designed in different ages of computing with different problems, hardware capabilities and needs.

## The right tool for the right job

For me, understanding a programming language, framework or stack implies love and fun. If you are not having fun and you don't love it, it might because you don't understand it well enough, or you're using it in a wrong way.

If they choose a screwdriver handler to hammer some nails, fine by me.

Convincing someone to jump from Java or Ruby to Go is not going to be achieved by comparing them. The influence by an example, diplomacy and the spread of your worldview would do a better job.

To give you an example, I have a good friend who is coding Rust these days. We came to agreement that Go and Rust are not competitors and found [Dave](https://dave.cheney.net/2015/07/02/why-go-and-rust-are-not-competitors) Cheney's opinion in the process. Last week I was presenting my friend with a few of Go advantages and he was impressed. Maybe he will give it a Go, maybe he wouldn't. That doesn't put us in disagreement, nor forcing him or me to choose sides.

Over the time, some of my colleagues asked me what I thought about some framework or language. I remember the discussion about TypeScript, that I had with one of them. My answer was simple : why should I learn a new language, when Javascript is enough for my needs? It's not about anything else, other than needs.

## Examples from History

* The year was 1964. [John Kemeny](https://en.wikipedia.org/wiki/John_G._Kemeny) and [Thomas Kurtz](https://en.wikipedia.org/wiki/Thomas_E._Kurtz) made their mission to allow non-expert users to interact with the computer. BASIC language was born - where B comes from Beginner. They had no language to compare to. Before BASIC, punched cards were top notch.

* A swiss computer scientist [Niklaus Wirth](https://en.wikipedia.org/wiki/Niklaus_Wirth) creates a bunch of computer languages, in his search for Pascal. He made it back in 1970, despite the fact that Americans pronounce its name as "nickel's worth". Again, the man hadn't needed to compare languages, mostly because C wasn't invented yet.

* Aged 31, [Dennis Ritchie](https://en.wikipedia.org/wiki/Dennis_Ritchie) creates the C language in which many of the language features "looked like a good thing to do". So, again, a man who didn't had to look somewhere else to get his ideas. It seems to me that all C like languages were build with the same goal of doing good things.

* While doing his PhD. in Cambridge, [Bjarne Stroustrup](https://en.wikipedia.org/wiki/Bjarne_Stroustrup) - in his own words - "invented C++, wrote its early definitions, and produced its first implementation... chose and formulated the design criteria for C++, designed all its major facilities, and was responsible for the processing of extension proposals in the C++ standards committee". My best guess is that came from necessity not similarities or dissimilarities.

* Dutch programmer [Guido van Rossum](https://en.wikipedia.org/wiki/Guido_van_Rossum) was looking for a "hobby" programming project that would keep him occupied during the week around Christmas. Two years later (1991), Python was born.

* It' already weird that most of the creators of programming languages come from Nordic countries, but it's a fact that [Rasmus Lerdorf](https://en.wikipedia.org/wiki/Rasmus_Lerdorf) a Danish-Canadian programmer creates PHP in 1995. Rasmus  did not intend the early PHP to become a new programming language.

* In 1995 [Brendan Eich](https://en.wikipedia.org/wiki/Brendan_Eich) created the first version of JavaScript in just ten days in order to accommodate the Navigator 2.0 Beta release schedule.

* [James Gosling](https://en.wikipedia.org/wiki/James_Gosling) invents Java for providing an alternative to the C++/C programming languages, as an internal project at Sun Microsystems. The keyword here is alternative.

## You Need Your Own Language?

If you feel like you need your own language or you want to have some fun, you can create it using a language that you know well and [PEG](https://en.wikipedia.org/wiki/Parsing_expression_grammar) or [ANTLR](https://en.wikipedia.org/wiki/ANTLR).

On the subject of "This has always bugged me and I can do better!", I've spent some time playing with [ANTLR4](https://github.com/antlr/antlr4) and created a grammar for a language called Greuceanu (in the Romanian mythology, Greuceanu is a young brave man who finds that the Sun and the Moon have been stolen by entities called `zmei`. After a long fight with the three `zmei` and their wives (`zmeoaice`), Greuceanu sets the Sun and the Moon free so the people on Earth have light again).

Indeed, writing the grammar for a new language takes time, but there is help for that too : [Visual Studio](https://marketplace.visualstudio.com/items?itemName=SamHarwell.ANTLRLanguageSupport), [Eclipse](https://marketplace.eclipse.org/content/antlr-4-ide) and [Intellij](https://plugins.jetbrains.com/plugin/7358-antlr-v4-grammar-plugin)
