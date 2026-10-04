import { ArrowLeft, ArrowRight, GithubLogo, Heart } from "@phosphor-icons/react";
import { isTauri } from "@tauri-apps/api/core";
import { type ReactNode, useEffect, useRef, useState } from "react";

import { Button } from "./ui/Button";

interface OnboardingPageProps {
  onExplore: () => void;
  onInstallExtension: () => void;
  onOpenGitHub: () => void;
  onOpenCoffee: () => void;
  onDonate: () => void;
  links: {
    extension: string;
    github: string;
    coffee: string;
    donate: string;
  };
}

const linkButtonClasses =
  "inline-flex items-center justify-center gap-2 rounded-full border border-border bg-surface text-sm font-semibold text-text-primary transition-colors hover:border-brand-300 hover:bg-surface-hover focus-visible:outline-2 focus-visible:outline-brand-500 focus-visible:outline-offset-2";

function ExternalLink({
  href,
  onOpen,
  className,
  children,
}: {
  href: string;
  onOpen: () => void;
  className: string;
  children: ReactNode;
}) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      onClick={(event) => {
        if (isTauri()) {
          event.preventDefault();
          onOpen();
        }
      }}
      className={className}
    >
      {children}
    </a>
  );
}

const SLIDES = [
  {
    title: "Catch the glitch. Keep the context.",
    description:
      "Record a browser issue. DAWG sanitizes the session and saves a portable artifact.",
    image: "/onboarding-workflow.svg",
    imageAlt: "Illustration of a browser capture becoming a portable DAWG artifact",
  },
  {
    title: "Replay the steps, not a video.",
    description:
      "Replay DOM changes and mouse activity to see what happened without a screen video.",
    image: "/onboarding-replay.svg",
    imageAlt: "Illustration of DOM events and mouse activity rebuilding a browser session",
  },
  {
    title: "Pass the finding from QA to Dev.",
    description: "Export a DAWG artifact so others can inspect and replay the same evidence.",
    image: "/onboarding-share.svg",
    imageAlt: "Illustration of a DAWG artifact shared from QA to development",
  },
  {
    title: "You're ready to dig in.",
    description: "Open the dashboard for a guided tour of the workspace. DAWG is open source.",
    image: "/onboarding-complete.svg",
    imageAlt: "DAWG logo surrounded by confetti",
  },
] as const;

type SlideActions = Pick<
  OnboardingPageProps,
  "onInstallExtension" | "onOpenGitHub" | "onOpenCoffee" | "links"
>;

function OnboardingSlide({
  index,
  onInstallExtension,
  onOpenGitHub,
  onOpenCoffee,
  links,
}: SlideActions & { index: number }) {
  const slide = SLIDES[index];

  return (
    <div className="grid min-h-96 md:grid-cols-[minmax(0,1fr)_minmax(0,1.05fr)]">
      <div className="flex flex-col justify-center px-7 py-8 sm:px-10 sm:py-10 lg:px-12">
        <h1
          aria-live="polite"
          className="max-w-lg text-3xl font-bold leading-[1.12] tracking-tight text-text-primary sm:text-4xl"
        >
          {slide.title}
        </h1>
        <p className="mt-4 max-w-md text-sm leading-6 text-text-secondary">{slide.description}</p>

        {index === 0 && (
          <ExternalLink
            href={links.extension}
            onOpen={onInstallExtension}
            className="mt-7 w-fit text-left text-sm font-semibold text-brand-600 underline decoration-brand-300 underline-offset-4 transition-colors hover:text-brand-700 focus-visible:rounded-sm focus-visible:outline-2 focus-visible:outline-brand-500 focus-visible:outline-offset-2"
          >
            Install the browser extension
          </ExternalLink>
        )}
        {index === 2 && (
          <p className="mt-7 text-xs font-medium text-text-tertiary">
            Review captured data before sharing an artifact.
          </p>
        )}
        {index === SLIDES.length - 1 && (
          <div className="mt-7 flex flex-wrap gap-3">
            <ExternalLink
              href={links.github}
              onOpen={onOpenGitHub}
              className={`${linkButtonClasses} h-9 px-4`}
            >
              <GithubLogo size={17} aria-hidden="true" />
              Slaviors Group
            </ExternalLink>
            <ExternalLink
              href={links.coffee}
              onOpen={onOpenCoffee}
              className={`${linkButtonClasses} h-9 px-4`}
            >
              <Heart size={17} aria-hidden="true" />
              Buy us a coffee
            </ExternalLink>
          </div>
        )}
      </div>

      <div className="onboarding-art-panel flex min-h-64 items-center justify-center overflow-hidden border-t border-border p-6 sm:p-9 md:min-h-0 md:border-t-0">
        {index === SLIDES.length - 1 ? (
          <div className="relative w-full max-w-130">
            <img src={slide.image} alt="" className="h-auto w-full object-contain" />
            <img
              src="/paw-dawg-mark.svg"
              alt={slide.imageAlt}
              className="absolute top-[48.75%] left-1/2 w-[36%] -translate-x-1/2 -translate-y-1/2 object-contain"
            />
          </div>
        ) : (
          <img
            src={slide.image}
            alt={slide.imageAlt}
            className="h-auto max-h-96 w-full max-w-130 object-contain"
          />
        )}
      </div>
    </div>
  );
}

export function OnboardingPage({
  onExplore,
  onInstallExtension,
  onOpenGitHub,
  onOpenCoffee,
  onDonate,
  links,
}: OnboardingPageProps) {
  const [slideIndex, setSlideIndex] = useState(0);
  const nextRef = useRef<HTMLButtonElement>(null);
  const isLastSlide = slideIndex === SLIDES.length - 1;

  useEffect(() => {
    nextRef.current?.focus();
  }, []);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (
        event.key !== "Enter" ||
        event.repeat ||
        event.isComposing ||
        event.altKey ||
        event.ctrlKey ||
        event.metaKey ||
        event.shiftKey
      )
        return;
      const target = event.target;
      if (
        target instanceof Element &&
        target.closest("button, a, input, select, textarea, [contenteditable='true']")
      )
        return;
      event.preventDefault();
      nextRef.current?.click();
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, []);

  const changeSlide = (nextIndex: number) => {
    if (nextIndex < 0 || nextIndex >= SLIDES.length) return;
    setSlideIndex(nextIndex);
  };

  const slideActions = { onInstallExtension, onOpenGitHub, onOpenCoffee, links };

  return (
    <main className="flex min-h-screen flex-col overflow-y-auto bg-canvas px-6 py-6 text-text-primary sm:px-10 sm:py-8">
      <header className="mx-auto flex w-full max-w-6xl items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <img src="/paw-dawg.svg" alt="" className="h-10 w-10 object-contain" />
          <div>
            <span className="block text-[15px] font-bold leading-tight tracking-tight">DAWG</span>
            <span className="block text-xs text-text-tertiary">Desktop workspace</span>
          </div>
        </div>
        <span className="rounded-full border border-border bg-surface px-3 py-1.5 text-xs font-semibold text-text-secondary">
          {slideIndex + 1} of {SLIDES.length}
        </span>
      </header>

      <div className="mx-auto flex w-full max-w-6xl flex-1 items-center py-6 sm:py-8">
        <div className="w-full overflow-hidden rounded-3xl border border-border bg-surface shadow-card">
          <OnboardingSlide index={slideIndex} {...slideActions} />
        </div>
      </div>

      <div className="mx-auto flex w-full max-w-6xl flex-wrap items-center justify-between gap-4 pb-2">
        <div className="flex items-center gap-3">
          <span className="sr-only" aria-live="polite">
            Slide {slideIndex + 1} of {SLIDES.length}
          </span>
          <div className="flex gap-1.5" aria-hidden="true">
            {SLIDES.map((item, index) => (
              <span
                key={item.image}
                className={`h-1.5 rounded-full ${index === slideIndex ? "w-7 bg-brand-500" : "w-3 bg-brand-200"}`}
              />
            ))}
          </div>
          {slideIndex > 0 && (
            <Button
              variant="ghost"
              onClick={() => changeSlide(slideIndex - 1)}
              iconLeft={<ArrowLeft size={16} />}
            >
              Back
            </Button>
          )}
        </div>
        <div className="flex flex-wrap items-center gap-3">
          {isLastSlide ? (
            <ExternalLink
              href={links.donate}
              onOpen={onDonate}
              className={`${linkButtonClasses} h-11 px-5`}
            >
              <Heart size={17} aria-hidden="true" />
              Donate
            </ExternalLink>
          ) : (
            <Button size="lg" variant="secondary" onClick={onExplore}>
              Skip intro
            </Button>
          )}
          <Button
            ref={nextRef}
            size="lg"
            onClick={isLastSlide ? onExplore : () => changeSlide(slideIndex + 1)}
            iconRight={<ArrowRight size={17} />}
          >
            {isLastSlide ? "Open dashboard" : "Next"}
          </Button>
        </div>
      </div>
    </main>
  );
}
