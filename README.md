# Adderscript
An easy-to-implement language for content programming. Don't use this yet, it's not done.

## About
Adderscript is a language intended for adding content programming functionality to a game (or non-game) project, such
as an authoritative backend server or a game client (or any non-networked game!) The idea is that generally speaking,
when developing games there is a lot of code that logically should block to make sense, but is not easily doable in
the language you're using.

Consider the following example of an RPG game where you interact with a door and it opens with a cool sequence. The
following is pseudocode for how it would more or less look:

```python
def use_door_lever(door):
    messagebox("You pull the lever...")
    do_animation("anim_pull_interact")
    sleep(2) # Wait for the animation to complete..
    sound("sound_rumbling")
    door.do_animation("anim_slideopen")
    sleep(4) # The door is opening..
    messagebox("A door opens!")
```

The above example, while written as Python code, is pseudocode for how one might handle a lever that opens a door,
including an animation, sound effects and some text. In most languages, it's not feasible to create a thread per 'script'
simply because when this runs for multiple entities (say every portal has its own script) you end up with lots of
threads.

Another option, present in some languages, are fibers. Those would be light-weight threads that aren't as resource-heavy
as regular threads. Still, they're generally not exactly what you intend to have since you now eventually still have an
X number of fibers running, and you're most likely introducing a new library anyway.

Yet another option, available in all languages, is writing a state machine. This can work exactly how you want it, without
having to spawn a new thread per instance, but the code you end up writing will be more verbose:

```java
void cycle() {
    switch (state) {
        case 0:
            messagebox("You pull the lever...");
            doAnimation("anim_pull_interact");
            state = 1;
            sleep(2); // Doesn't sleep the thread, but instructs your runtime to continue this instance after 2s.
            break;
        case 1:
            sound("sound_rumbling");
            door.doAnimation("anim_slideopen");
            sleep(4);
            break;
        // and so on
    }
}
```

This Java code example would be a single state machine. For this single case it may look fine, but in the end you
will have to write a class per script, write a state machine for your logic (which will become hard to maintain for
more complex scripts) and eventually end up with a lot of boilerplate.

It can be simpler, though. Using Adderscript:

```c
// Declare a listener function, identify the event with target "obj_door". Takes one argument: the interacted obj.
on object_interact("obj_door")(gameobj door) {
    messagebox("You pull the lever...");
    do_animation("anim_pull_interact");
    sleep(2);
    sound("sound_rumbling");
    door.do_animation("anim_slideopen");
    sleep(4);
    messagebox("A door opens!");
}
```

Now we have code that looks clean, is maintainable, and best of all, doesn't even require recompilation when you
have to make a change!

The downside is that you're now introducing a new language to your project.. but how bad can it be? After all, with
Adderscript you define the runtime yourself. Adder provides no runtime at all, nor any default listeners. This means
that Adderscript as a language is not standalone usable. It spits out bytecode that is very easy to use, though, and
execution engines are available for a fair amount of libraries. All you have to do is define functions that fit your
specific use case, such as interactions, UI functions or other things.

## Roadmap
 - Method references
 - Anonymous functions
 - Structs
 - Annotations?
