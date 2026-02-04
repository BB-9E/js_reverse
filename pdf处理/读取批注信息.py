import aspose.pdf as ap
from aspose.pdf.devices import PngDevice, Resolution
from aspose.pdf.annotations import AnnotationFlags


def export_annotations_to_images(pdf_path, output_prefix="annotation"):
    doc = ap.Document(pdf_path)
    img_index = 1

    for page in doc.pages:
        page_num = page.number

        for ann in page.annotations:

            rect = ann.rect
            if not rect:
                continue

            # ===== 隐藏当前页中“非当前”的批注 =====
            for other in page.annotations:
                if other.name != ann.name:
                    other.flags |= AnnotationFlags.HIDDEN
                else:
                    other.flags &= ~AnnotationFlags.HIDDEN

            # ===== 裁剪到当前批注区域 =====
            original_crop = page.crop_box
            padding = 20
            crop_rect = ap.Rectangle(
                rect.llx - padding,
                rect.lly - padding,
                rect.urx + padding,
                rect.ury + padding,
                True
            )
            page.crop_box = crop_rect

            # ===== 导出图片 =====
            resolution = Resolution(200)
            device = PngDevice(resolution)
            filename = f"{output_prefix}_p{page_num}_{img_index}.png"
            device.process(page, filename)
            print(f"已生成: {filename}")

            # ===== 恢复页面状态 =====
            page.crop_box = original_crop
            for other in page.annotations:
                other.flags &= ~AnnotationFlags.HIDDEN

            img_index += 1


if __name__ == '__main__':
    export_annotations_to_images("./1.pdf")
