+++
title = "Building a Filesystem to Actually Understand One"
date = "2025-12-16"
author = "kiiNODA"
+++
I've spent years thinking I know about filesystems. What an inode is. What the superblock is. But there's a difference between knowing something theoretically and understanding it through your fingers - the kind of understanding that comes from making something crash, fixing it, and seeing exactly why it works now.

Most people don't learn for the sheer pleasure of learning anymore. I say this without judgment, but I notice it. The pragmatic view dominates: learn what you need for the next sprint, the next promotion, the next job change. But I have gaps in knowledge and I'm still hungry to fill them, even when there's no career ROI. Filesystems were one of those gaps - something I could describe on paper but had never actually built.

So I decided to build one. In Python. From scratch.

## Why Another Toy Filesystem?

Every operating systems course teaches filesystems through textbooks. You study the theory: inodes store metadata, directory entries map names to inodes, bitmaps track free space. You might trace through ext4 or NTFS diagrams. But do you ever implement one yourself, watching the raw bytes take shape on disk?

I wanted that experience. Not to build something production-ready - that would be insane - but to make the abstract concepts concrete. So I'm building SimpleFS with intentional simplifications: 4KB maximum file size, no permissions, no symbolic links, an implementation small enough to understand completely in an evening. It's a proof of concept designed purely for learning.

![SimpleFS Layout](/assets/building-a-filesystem--layout.png)

The CLI lets you run familiar commands: `init`, `touch`, `mkdir`, `write`, `read`, `cp`, `mv`, `rm`. Then you can immediately hexdump the raw disk image to see your data in binary. Create a file, hexdump block 0, and there's the superblock magic bytes. Allocate an inode, hexdump block 1, and there's bit 0 flipped to 1 in the inode bitmap.

Theory became tangible.

## The Journey So Far

I structured this as 10 progressive sessions, each designed for instant gratification. Every session produces working, visible results. Dopamine hits while building. This matters more than people admit - learning needs momentum.

**First: the bootstrap**. Implement `init` to create a fresh filesystem. I got to see the filesystem taking shape by dumping raw bytes. I did this on an evening between checking news about Bucharest's mayoral election, and watching those magic bytes appear felt like actual creation.

**Next: block devices and bitmaps**. This is where theory met practice hard. Block devices seemed straightforward - wrap file I/O to work with fixed-size chunks instead of arbitrary bytes. But implementing it forced me to understand *why* filesystems think in blocks. Disk sectors and memory pages are aligned to specific sizes. Working with block numbers is simpler than calculating byte offsets every time. It's not just convention, it's physics.

Then came bitmaps for tracking free space. I thought I'd need just one bitmap, but no - filesystems need two. One for file metadata (inodes), one for actual disk blocks. They track different resources with different pool sizes. Some blocks get allocated for filesystem structures that don't correspond to files at all. The conceptual separation matters.

And then implementing bitmap operations pushed me into bit manipulation, where I wrote code to set a bit and my tests failed because I used simple assignment instead of bitwise OR. I was accidentally erasing other allocations in the same byte. Small detail. The difference between correctness and subtle corruption.

**Then: integration and statistics**. I tied the block device and bitmap system together, integrated them into the filesystem init logic, and added a `stat` command to read the superblock and display filesystem statistics. The infrastructure was starting to cohere into something functional.

**After that: verification**. I could now hexdump specific byte offsets and see the exact bit patterns I expected. Block 0 shows the magic bytes. Block 1 has bit 0 set for the root inode. Block 2 has the first 67 bits set for reserved blocks. I saw `0xff 0xff 0xff 0xff 0xff 0xff 0xff 0xff 0x07` at offset 0x2000 (8192) and recognized it as binary for "blocks 0-66 are allocated."

![Hexdump showing filesystem metadata](/assets/building-a-filesystem--hexdump.png)

The filesystem metadata became real, visible, tangible.

**Most recently: inodes**. Every file on your computer has metadata: size, type, where its data lives on disk. Filesystems store this in structures called inodes. The key insight was understanding the layers of indirection. When you allocate an inode, you're doing three separate things: asking the bitmap for a free number, creating an in-memory structure with metadata, then writing that structure to a specific location on disk.

The inode number isn't just an ID - it's a coordinate system. Each number maps to a specific block and byte offset, calculated mathematically from the number itself. The math maps logical concept to physical location.

Then I hit the read-modify-write pattern that's everywhere in filesystem code. You can't just write 256 bytes of inode data at an arbitrary offset - you have to read the entire 4KB block, modify the relevant slice, then write the whole block back. Blocks are the atomic unit of disk I/O.

And I learned why you need mutable byte arrays before modifying - Python's immutable bytes throw errors when you try to update slices, so you have to convert first. Small language detail with huge practical implications.

The tests now prove the round trip works: allocate an inode, modify its metadata, write it to disk, read it back, verify the fields match.

## What I Actually Learned

Every time I thought I understood something from reading about it, implementing it showed me I was missing a layer. Sometimes multiple layers.

I learned that alignment isn't just optimization, it's correctness. I learned that bit manipulation bugs are silent until they corrupt something three operations later. I learned that the read-modify-write pattern exists because hardware demands it, not because software engineers love extra work.

I learned that for all my years of running `ls` and wondering how it finds files, making it tangible requires getting your hands dirty with byte offsets and serialization formats. The understanding is different. Deeper. Permanent.

I'm not doing this for a project at work. I'm not building a startup. I'm not even sure anyone will read this. I'm doing it purely for the love of learning and making OS concepts concrete rather than abstract.

Still building. Still learning. Six more sessions to go.
