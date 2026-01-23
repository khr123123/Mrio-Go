# Super Mario-like Game (Go)

A personal 2D platformer game project inspired by **Super Mario**, developed in **Go**.
This project focuses on learning game loop design, real-time rendering, entity management, and basic game architecture.

---

## 📌 Project Overview

* **Language**: Go
* **Game Type**: 2D Side-Scrolling Platformer
* **Purpose**:

    * Practice game development using Go
    * Learn game loop control, entity updates, and collision concepts
    * Build a scalable structure for future features

---

## 🧩 Current Features (Implemented)

The following core systems are **already implemented**:

* ✅ **Background System**

    * Static background rendering
    * Screen size and camera baseline setup

* ✅ **Player Control**

    * Basic movement control
    * Player state updates handled in the game loop

* ✅ **Monster System**

    * Monster spawning (randomized or fixed positions)
    * Monster refresh/update logic
    * Basic lifecycle management

* ✅ **Game Loop**

    * Frame-based update and render cycle
    * Centralized update control

---

## 🚧 Features In Progress / Not Yet Implemented

These features are **planned but not completed**:

### 🎮 Gameplay Mechanics

* ⬜ Player jump physics (gravity, acceleration)
* ⬜ Collision detection (player ↔ ground / monsters)
* ⬜ Player damage & death logic
* ⬜ Monster AI (patrol, chase, simple behaviors)

### 🗺️ World & Level Design

* ⬜ Tile-based map system
* ⬜ Multiple levels / stages
* ⬜ Scrolling camera
* ⬜ Platforms, obstacles, and pipes

### 🎨 Visual & Audio

* ⬜ Sprite animations
* ⬜ Player & monster animation states
* ⬜ Sound effects (jump, hit, death)
* ⬜ Background music

### 🧠 Game State Management

* ⬜ Start menu
* ⬜ Pause system
* ⬜ Game over screen
* ⬜ Restart logic

---

## 🧱 Suggested Project Structure

```text
/asset
  /images
  /sounds
/entity
  player.go
  monster.go
/game
  loop.go
  state.go
/world
  background.go
  level.go
main.go
```

---

## 📋 Development Checklist (Task Management)

### Phase 1: Core Mechanics

* [ ] Gravity & jump system
* [ ] Collision detection
* [ ] Basic physics tuning

### Phase 2: Gameplay Expansion

* [ ] Monster behavior patterns
* [ ] Player health system
* [ ] Scoring system

### Phase 3: Content & Polish

* [ ] Level editor or data-driven maps
* [ ] Animations
* [ ] Sound & music
* [ ] UI improvements

---

## 🎯 Development Principles

* Keep systems **simple and modular**
* Avoid premature optimization
* Prefer clarity over complexity
* Refactor when patterns become clear

---

## 📝 Notes

This project is under active development.
The goal is not to perfectly clone Mario, but to **understand how a platformer works from scratch** using Go.

---

## 📈 Future Goals

* Save/load game state
* Configurable difficulty
* More enemy types
* Boss fights
* Performance optimization

---

**Author**: Personal Learning Project
**Status**: Early Development 🚀
