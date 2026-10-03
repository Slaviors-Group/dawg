import { FileText, Folder, GridFour, Image, PlayCircle } from "@phosphor-icons/react";
import type { ReactNode } from "react";

interface EmptyStateProps {
  icon?: ReactNode;
  title: string;
  description?: string;
  action?: ReactNode;
}

export function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="empty-state">
      <div className="empty-state__scene" aria-hidden="true">
        <span className="empty-state__ring empty-state__ring--outer" />
        <span className="empty-state__ring empty-state__ring--middle" />
        <span className="empty-state__ring empty-state__ring--inner" />

        <span className="empty-state__orbit-icon empty-state__orbit-icon--grid-left">
          <GridFour size={22} weight="regular" />
        </span>
        <span className="empty-state__orbit-icon empty-state__orbit-icon--file-top empty-state__orbit-icon--tile">
          <FileText size={25} weight="regular" />
        </span>
        <span className="empty-state__orbit-icon empty-state__orbit-icon--grid-right empty-state__orbit-icon--tile">
          <GridFour size={27} weight="regular" />
        </span>
        <span className="empty-state__orbit-icon empty-state__orbit-icon--folder empty-state__orbit-icon--tile">
          <Folder size={28} weight="regular" />
        </span>
        <span className="empty-state__orbit-icon empty-state__orbit-icon--file-bottom">
          <FileText size={21} weight="regular" />
        </span>
        <span className="empty-state__orbit-icon empty-state__orbit-icon--image">
          <Image size={23} weight="regular" />
        </span>
        <span className="empty-state__orbit-icon empty-state__orbit-icon--play">
          <PlayCircle size={24} weight="regular" />
        </span>

        <span className="empty-state__document">
          <span className="empty-state__main-icon">{icon ?? <FileText size={34} />}</span>
        </span>
      </div>

      <div className="empty-state__copy">
        <h3 className="empty-state__title">{title}</h3>
        {description && <p className="empty-state__description">{description}</p>}
      </div>
      {action && <div className="empty-state__action">{action}</div>}
    </div>
  );
}
